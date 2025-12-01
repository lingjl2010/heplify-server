COMMANDS=$(patsubst cmd/%,%,$(wildcard cmd/*))
TARGETS=$(foreach cmd,${COMMANDS},cmd/${cmd}/${cmd})
COMMIT?=$(shell git rev-parse --short=8 HEAD)$(shell git diff --quiet || echo '-dirty')
BASE?=csighub.tencentyun.com/tccc/base:latest
DOCKER_BUILD_TARGETS=
DOCKER_PUSH_TARGETS=
SOFTPBX_BUILD_TARGETS=
SOFTPBX_PUSH_TARGETS=
SOFTPBX_VPC_BUILD_TARGETS=
SOFTPBX_VPC_PUSH_TARGETS=
export LD_LIBRARY_PATH=${CURDIR}/docker
# export TCCC_DOCKER_LOCK=${CURDIR}/.lock
ifeq ($(OS),Windows_NT)
	EXE_SUFFIX=.exe
else
	EXE_SUFFIX=
endif
GOVERSION=$(shell go version|grep -o 'go[0-9][0-9.]\+'|cut -c 3-)
GOLANGCILINTVERSION=v1.64.5
# go1.24开始会缓存go run指定版本的结果，因此不需要额外保存binary了 这里不使用go tool的原因是linter的依赖太多，容易影响生产代码的依赖
STDLINT=go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINTVERSION)
TENCENTLINT=./tencentlint/docker/custom-gcl$(EXE_SUFFIX)
CUSTOMLINT=./custom-gcl$(EXE_SUFFIX)
GOEXPERIMENTENV=GOEXPERIMENT=synctest

all:
#	 $(MAKE) prepare
	go build ./...

heplify-server:
	cd softpbx/heplify-server && go build ./...

prepare: # i18n的代码生成依赖源码成功构建，因此需要两步generate
	go generate -x -skip "i18n" ./...
	go generate -x -run "i18n" ./...

test:
	$(GOEXPERIMENTENV) go test -vet=off -covermode=atomic -coverprofile=cover.out -shuffle=on -timeout 20m -gcflags=all=-l ./...

test-case-by-case:
	./go-test-case-by-case.sh

short-test:
	$(GOEXPERIMENTENV) go test -short -vet=off -covermode=atomic -coverprofile=cover.out -shuffle=on -gcflags=all=-l ./...

clean:
	rm -f ${TARGETS} *.list

vulncheck:
	go tool govulncheck -show verbose ./...

lint-install: $(TENCENTLINT) $(CUSTOMLINT)

$(TENCENTLINT):
	cd tencentlint/docker && sed 's/version: .*/version: $(GOLANGCILINTVERSION)/' -i .custom-gcl.yml && $(STDLINT) custom -v

$(CUSTOMLINT):
	sed 's/version: .*/version: $(GOLANGCILINTVERSION)/' -i .custom-gcl.yml && $(STDLINT) custom -v

define CHECK_LINT_VERSION
lint-version-check-$(1): $(1)
	@VERSION=$$$$(go version -m $(1)|grep -o 'go[0-9][0-9.]\+'|cut -c 3-); \
	LINTVERSION=$$$$($(1) --version|grep -o 'v[0-9.]\+\(-custom-gcl\)\?'); \
	LINTVERSION=$$$${LINTVERSION%-custom-gcl} ; \
	if [ "$$$$VERSION" != "$(GOVERSION)" -o "$$$$LINTVERSION" != "$(GOLANGCILINTVERSION)" ]; then \
		echo "$(1) go/lint version mismatch: $$$$VERSION != $(GOVERSION)/$$$$LINTVERSION != $(GOLANGCILINTVERSION) , reinstalling"; \
		rm -f $(1); \
		$(MAKE) $(1); \
	fi

endef
$(eval \
$(call CHECK_LINT_VERSION,$(TENCENTLINT))\
$(call CHECK_LINT_VERSION,$(CUSTOMLINT))\
)
lint-version-check: lint-version-check-$(TENCENTLINT) lint-version-check-$(CUSTOMLINT)

lint: $(TENCENTLINT) $(CUSTOMLINT)
	$(MAKE) lint-version-check
	go mod tidy -diff
	$(GOEXPERIMENTENV) $(TENCENTLINT) run -c ./tencentlint/.golangci.yml -v --new-from-rev origin/master
	$(GOEXPERIMENTENV) $(CUSTOMLINT) run -v --new-from-rev origin/master
	git -C tencentlint checkout .

lint-fix: $(TENCENTLINT) $(CUSTOMLINT)
	$(MAKE) lint-version-check
	go mod tidy
	-$(GOEXPERIMENTENV) $(TENCENTLINT) run -c ./tencentlint/.golangci.yml -v --new-from-rev origin/master --fix
	-$(GOEXPERIMENTENV) $(CUSTOMLINT) run -v --new-from-rev origin/master --fix

sync-rainbow:
	# 需要指定操作人
	$(MAKE) -C rainbow
	docker run --rm -v "$(CURDIR):/data" -w /data -ePLUGIN_USER -ePLUGIN_APPID=96dab1f5-850e-4823-8696-32829bcd8050 -ePLUGIN_FULL=true csighub.tencentyun.com/okhowang/rainbow_gitops:latest

cmd/%:
	cd $(dir $@) && go build

docker-build-base:
	docker build --pull -t csighub.tencentyun.com/tccc/base:latest -f docker/base.ci.Dockerfile .

# 声明本地构建的模块 产生target docker-build-$(1) docker-push-$(1)
# docker-build-$(1) 用于构建镜像 在ci中使用 使用本地已经build过的产物节约构建时间
# docker-push-$(1) 推送docker-build-$(1)的镜像到仓库
define DOCKER_RULE
docker-build-$(1): cmd/$(1)/$(1)
	$$(eval TMP:=$$(shell mktemp -d))
	cp --parents -r docker cmd/$(1)/$(1) $$(TMP)
	if ls "cmd/$(1)"|grep '\.sh$$$$'; then cp -v --parents cmd/$(1)/*.sh $$(TMP); fi
	if [ -e $(1)/conf ]; then cp --parents -r $(1)/conf $$(TMP); fi
	docker build --pull -t csighub.tencentyun.com/tccc/$(1):$$(COMMIT) -f docker/$(2).ci.Dockerfile --build-arg COMMAND=$(1) --build-arg BASE=$$(BASE) $$(TMP)
	docker tag csighub.tencentyun.com/tccc/$(1):$$(COMMIT) csighub.tencentyun.com/tccc/$(1):latest
	rm -rf $$(TMP)

DOCKER_BUILD_TARGETS+=docker-build-$(1)

docker-push-$(1):
	docker push csighub.tencentyun.com/tccc/$(1):$$(COMMIT)
	docker push csighub.tencentyun.com/tccc/$(1):latest

DOCKER_PUSH_TARGETS+=docker-push-$(1)

endef
$(eval \
$(call DOCKER_RULE,report-webserver,general)\
$(call DOCKER_RULE,admin-loginserver,going)\
$(call DOCKER_RULE,ccc-wx,going)\
$(call DOCKER_RULE,ccc-omni-channel,general)\
$(call DOCKER_RULE,statistic-tools,tool)\
$(call DOCKER_RULE,statistic-timer,general)\
$(call DOCKER_RULE,callback-proxy,general)\
$(call DOCKER_RULE,interface,general)\
$(call DOCKER_RULE,acd-server,general)\
$(call DOCKER_RULE,qidian-server,going)\
$(call DOCKER_RULE,qidian-sync,going)\
$(call DOCKER_RULE,acd-timer,general)\
$(call DOCKER_RULE,acd-tools,tool)\
$(call DOCKER_RULE,state-server,general)\
$(call DOCKER_RULE,state-cleaner,general)\
$(call DOCKER_RULE,predictive-scheduler,general)\
$(call DOCKER_RULE,cruise-scheduler,general)\
$(call DOCKER_RULE,websocket-interface,general)\
$(call DOCKER_RULE,kafka-consumer,going)\
$(call DOCKER_RULE,media-server,general)\
$(call DOCKER_RULE,session-server,general)\
$(call DOCKER_RULE,servnum-control,general)\
$(call DOCKER_RULE,admin-webserver,going)\
$(call DOCKER_RULE,admin-looper,general)\
$(call DOCKER_RULE,admin-callout-stat-job,tool)\
$(call DOCKER_RULE,admin-cost-job,tool)\
$(call DOCKER_RULE,admin-pstn-number-collector,tool)\
$(call DOCKER_RULE,admin-number-checker,tool)\
$(call DOCKER_RULE,admin-number-recycle,tool)\
$(call DOCKER_RULE,syncpstnconfig,tool)\
$(call DOCKER_RULE,demo-interface-invocation-server,general)\
$(call DOCKER_RULE,logpusher,tool)\
$(call DOCKER_RULE,seat-data-statistics,tool)\
$(call DOCKER_RULE,data-sync,tool)\
$(call DOCKER_RULE,asr-sample-tool,tool)\
$(call DOCKER_RULE,softpbx-sidecar,general)\
$(call DOCKER_RULE,softpbx-initer,tool)\
$(call DOCKER_RULE,opensipssyncer,general)\
$(call DOCKER_RULE,sipgateway,general)\
$(call DOCKER_RULE,capi-proxy,general)\
$(call DOCKER_RULE,synclicense-job,tool)\
$(call DOCKER_RULE,synccontract-job,tool)\
$(call DOCKER_RULE,acd-checker-job,tool)\
$(call DOCKER_RULE,init-eip-direct,tool)\
$(call DOCKER_RULE,scam-server,general)\
)

cmd/softpbx-sidecar/softpbx-sidecar:
	cd cmd/softpbx-sidecar && go build -tags=collector

# 声明一些自定义的开源模块 使用自己的Dockerfile 产生target docker-build-$(1) docker-push-$(1)
# docker-build-$(1) 用于构建镜像
# docker-push-$(1) 推送docker-build-$(1)的镜像到仓库
define DOCKER_RULE_SOFTPBX
docker-build-$(1):
	docker build --pull -t csighub.tencentyun.com/tccc/$(1):$$(COMMIT) -f softpbx/$(1)/Dockerfile softpbx/$(1)
	docker tag csighub.tencentyun.com/tccc/$(1):$$(COMMIT) csighub.tencentyun.com/tccc/$(1):latest

SOFTPBX_BUILD_TARGETS+=docker-build-$(1)

docker-push-$(1):
	docker push csighub.tencentyun.com/tccc/$(1):$$(COMMIT)
	docker push csighub.tencentyun.com/tccc/$(1):latest

SOFTPBX_PUSH_TARGETS+=docker-push-$(1)

endef
$(eval \
$(call DOCKER_RULE_SOFTPBX,rtp-server)\
$(call DOCKER_RULE_SOFTPBX,sbc-server)\
$(call DOCKER_RULE_SOFTPBX,sip-server)\
$(call DOCKER_RULE_SOFTPBX,uac-server)\
$(call DOCKER_RULE_SOFTPBX,reg-server)\
$(call DOCKER_RULE_SOFTPBX,pstn-proxy)\
)

define DOCKER_RULE_SOFTPBX_VPC
docker-build-$(1):
	docker build --pull -t ccr.ccs.tencentyun.com/tccc-images/$(1):$$(COMMIT) -f softpbx/$(1)/Dockerfile softpbx/$(1)
	docker tag ccr.ccs.tencentyun.com/tccc-images/$(1):$$(COMMIT) ccr.ccs.tencentyun.com/tccc-images/$(1):latest

SOFTPBX_VPC_BUILD_TARGETS+=docker-build-$(1)

docker-push-$(1):
	docker push ccr.ccs.tencentyun.com/tccc-images/$(1):$$(COMMIT)
	docker push ccr.ccs.tencentyun.com/tccc-images/$(1):latest

SOFTPBX_VPC_PUSH_TARGETS+=docker-push-$(1)

endef
$(eval \
$(call DOCKER_RULE_SOFTPBX_VPC,sip-proxy)\
$(call DOCKER_RULE_SOFTPBX_VPC,sip-proxy-rtp)\
)

define DOCKER_RULE_V2

docker-build-$(1)-v2:
	$$(eval TMP:=$$(shell mktemp -d))
	cp --parents -r . $$(TMP)
	docker build --pull -t csighub.tencentyun.com/tccc/$(1):$$(COMMIT) -f docker/$(2).ci.v2.Dockerfile --build-arg COMMAND=$(1) --build-arg BASE=$$(BASE) $$(TMP)
	docker tag csighub.tencentyun.com/tccc/$(1):$$(COMMIT) csighub.tencentyun.com/tccc/$(1):latest
	rm -rf $$(TMP)

endef

$(eval \
$(call DOCKER_RULE_V2,admin-pstn-number-collector,tool)\
$(call DOCKER_RULE_V2,admin-number-checker,tool)\
$(call DOCKER_RULE_V2,admin-number-recycle,tool)\
$(call DOCKER_RULE_V2,admin-webserver,going)\
)

all-bin: ${TARGETS}

all-docker-build: ${DOCKER_BUILD_TARGETS}

all-docker-push: ${DOCKER_PUSH_TARGETS}

softpbx-docker-build: ${SOFTPBX_BUILD_TARGETS}

softpbx-docker-push: ${SOFTPBX_PUSH_TARGETS}

softpbx-vpc-docker-build: ${SOFTPBX_VPC_BUILD_TARGETS}

softpbx-vpc-docker-push: ${SOFTPBX_VPC_PUSH_TARGETS}

# 公司内部包暂时忽略
# github.com/golangci/plugin-module-register/register lint工具包不分发 暂时忽略
# github.com/richardlehane/msoleps/types excelize依赖 误报忽略 实际为Apache2.0协议
# github.com/discoviking/fsm 被gosip依赖 暂时仅用于sip-gateway trtc公有云环境 暂时忽略
# bitbucket.org/ausocean/av/codec/pcm 被media-server依赖 暂时用于trtc公有云环境 暂时忽略
IGNORED_LICENSE=--ignore git.woa.com --ignore git.code.oa.com --ignore github.com/tencentcloud \
                		--ignore github.com/richardlehane/msoleps/types \
                		--ignore github.com/golangci/plugin-module-register/register \
                		--ignore github.com/discoviking/fsm \
                		--ignore bitbucket.org/ausocean/av/codec/pcm

check-license:
	go install github.com/google/go-licenses@latest
	go-licenses check ${IGNORED_LICENSE} \
		--log_file licenses-error.log --logtostderr=false \
        --disallowed_types forbidden,restricted,unknown \
        ./...

save-license:
	rm -rf license
	go install github.com/google/go-licenses@latest
	go-licenses save --save_path license ${IGNORED_LICENSE} \
		./...

.PHONY: all test clean all-bin lint lint-fix lint-install heplify-server
