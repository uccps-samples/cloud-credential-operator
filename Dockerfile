FROM openshift/golang-builder@sha256:4820580c3368f320581eb9e32cf97aeec179a86c5749753a14ed76410a293d83 AS builder
ENV __doozer=update BUILD_RELEASE=202202160023.p0.gaa55102.assembly.stream BUILD_VERSION=v4.10.0 OS_GIT_MAJOR=4 OS_GIT_MINOR=10 OS_GIT_PATCH=0 OS_GIT_TREE_STATE=clean OS_GIT_VERSION=4.10.0-202202160023.p0.gaa55102.assembly.stream SOURCE_GIT_TREE_STATE=clean 
ENV __doozer=merge OS_GIT_COMMIT=aa55102 OS_GIT_VERSION=4.10.0-202202160023.p0.gaa55102.assembly.stream-aa55102 SOURCE_DATE_EPOCH=1643142585 SOURCE_GIT_COMMIT=aa5510254782d506f59a0578661f22a8d639dc85 SOURCE_GIT_TAG=aa551025 SOURCE_GIT_URL=https://github.com/uccps-samples/cloud-credential-operator 
WORKDIR /go/src/github.com/uccps-samples/cloud-credential-operator
COPY . .
ENV GO_PACKAGE github.com/uccps-samples/cloud-credential-operator
RUN go build -ldflags "-X $GO_PACKAGE/pkg/version.versionFromGit=$(git describe --long --tags --abbrev=7 --match 'v[0-9]*')" ./cmd/cloud-credential-operator
RUN go build -ldflags "-X $GO_PACKAGE/pkg/version.versionFromGit=$(git describe --long --tags --abbrev=7 --match 'v[0-9]*')" ./cmd/ccoctl

FROM openshift/ose-base:v4.10.0.20220216.010142
ENV __doozer=update BUILD_RELEASE=202202160023.p0.gaa55102.assembly.stream BUILD_VERSION=v4.10.0 OS_GIT_MAJOR=4 OS_GIT_MINOR=10 OS_GIT_PATCH=0 OS_GIT_TREE_STATE=clean OS_GIT_VERSION=4.10.0-202202160023.p0.gaa55102.assembly.stream SOURCE_GIT_TREE_STATE=clean 
ENV __doozer=merge OS_GIT_COMMIT=aa55102 OS_GIT_VERSION=4.10.0-202202160023.p0.gaa55102.assembly.stream-aa55102 SOURCE_DATE_EPOCH=1643142585 SOURCE_GIT_COMMIT=aa5510254782d506f59a0578661f22a8d639dc85 SOURCE_GIT_TAG=aa551025 SOURCE_GIT_URL=https://github.com/uccps-samples/cloud-credential-operator 
COPY --from=builder /go/src/github.com/uccps-samples/cloud-credential-operator/cloud-credential-operator /usr/bin/
COPY --from=builder /go/src/github.com/uccps-samples/cloud-credential-operator/ccoctl /usr/bin/
COPY manifests /manifests
# Update perms so we can copy updated CA if needed
RUN chmod -R g+w /etc/pki/ca-trust/extracted/pem/
# TODO make path explicit here to remove need for ENTRYPOINT
# https://github.com/uccps-samples/installer/blob/a8ddf6619794416c4600a827c2d9284724d382d8/data/data/bootstrap/files/usr/local/bin/bootkube.sh.template#L347
ENTRYPOINT [ "/usr/bin/cloud-credential-operator" ]

LABEL \
        io.openshift.release.operator="true" \
        name="openshift/ose-cloud-credential-operator" \
        com.redhat.component="ose-cloud-credential-operator-container" \
        io.openshift.maintainer.product="OpenShift Container Platform" \
        io.openshift.maintainer.component="Cloud Credential Operator" \
        release="202202160023.p0.gaa55102.assembly.stream" \
        io.openshift.build.commit.id="aa5510254782d506f59a0578661f22a8d639dc85" \
        io.openshift.build.source-location="https://github.com/uccps-samples/cloud-credential-operator" \
        io.openshift.build.commit.url="https://github.com/uccps-samples/cloud-credential-operator/commit/aa5510254782d506f59a0578661f22a8d639dc85" \
        version="v4.10.0"

