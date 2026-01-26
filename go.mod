module github.com/aws/eks-anywhere

go 1.24.0

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/aws/aws-sdk-go v1.50.36
	github.com/aws/aws-sdk-go-v2 v1.30.1
	github.com/aws/aws-sdk-go-v2/config v1.26.6
	github.com/aws/aws-sdk-go-v2/credentials v1.17.7
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.167.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.27.4
	github.com/aws/eks-anywhere-packages v0.4.5
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice v0.0.0-00010101000000-000000000000
	github.com/aws/eks-distro-build-tooling/release v0.0.0-20211103003257-a7e2379eae5e
	github.com/aws/etcdadm-bootstrap-provider v1.0.17
	github.com/aws/etcdadm-controller v1.0.25
	github.com/aws/smithy-go v1.20.3
	github.com/bmc-toolbox/bmclib/v2 v2.1.1-0.20231206130132-1063371b9ed6
	github.com/docker/cli v27.0.3+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.13.0
	github.com/go-jose/go-jose/v3 v3.0.4
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/zapr v1.3.0
	github.com/gocarina/gocsv v0.0.0-20220304222734-caabc5f00d30
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v35 v35.3.0
	github.com/google/uuid v1.6.0
	github.com/nutanix-cloud-native/cluster-api-provider-nutanix v1.3.2
	github.com/nutanix-cloud-native/prism-go-client v0.3.4
	github.com/onsi/gomega v1.38.2
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.10.1
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/tinkerbell/tink v0.12.2
	github.com/vmware/govmomi v0.51.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.45.0
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa
	golang.org/x/net v0.47.0
	golang.org/x/oauth2 v0.33.0
	golang.org/x/text v0.31.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.18.5
	k8s.io/api v0.34.2
	k8s.io/apimachinery v0.34.2
	k8s.io/apiserver v0.34.2
	k8s.io/client-go v0.34.2
	k8s.io/component-base v0.34.2
	k8s.io/klog/v2 v2.130.1
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397
	oras.land/oras-go/v2 v2.6.0
	sigs.k8s.io/cluster-api v1.6.2
	sigs.k8s.io/cluster-api-provider-cloudstack v0.4.9-rc7
	sigs.k8s.io/cluster-api-provider-vsphere v1.14.0
	sigs.k8s.io/cluster-api/test v1.11.1
	sigs.k8s.io/controller-runtime v0.22.4
	sigs.k8s.io/yaml v1.6.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.26.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.15 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.2 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Jeffail/gabs/v2 v2.7.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/ProtonMail/go-crypto v1.1.3 // indirect
	github.com/VictorLowther/simplexml v0.0.0-20180716164440-0bff93621230 // indirect
	github.com/VictorLowther/soap v0.0.0-20150314151524-8e36fca84b22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.4
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/internal/configsources v0.0.0-00010101000000-000000000000 // indirect
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/internal/endpoints/v2 v2.0.0-00010101000000-000000000000 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmc-toolbox/common v0.0.0-20230717121556-5eb9915a8a5a // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/containerd/containerd v1.7.29 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gobuffalo/flect v1.0.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.17
	github.com/jacobweinstock/iamt v0.0.0-20230502042727-d7cdbe67d9ef // indirect
	github.com/jacobweinstock/registrar v0.4.7 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/stmcginnis/gofish v0.15.1-0.20231121142100-22a60a77be91 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sync v0.18.0
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.72.3 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/apiextensions-apiserver v0.34.2
	k8s.io/cluster-bootstrap v0.34.2 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	k8s.io/kubelet v0.29.5
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
)

replace (
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/internal/configsources => ./internal/aws-sdk-go-v2/internal/configsources
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/internal/endpoints/v2 => ./internal/aws-sdk-go-v2/internal/endpoints/v2
	github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice => ./internal/aws-sdk-go-v2/service/snowballdevice

	github.com/nutanix-cloud-native/cluster-api-provider-nutanix => ./internal/thirdparty/capx

	// need the modifications eksa made to the capi api structs
	sigs.k8s.io/cluster-api => github.com/snarkychef/cluster-api v0.1.0

	// Temporary until capc is updated to use a newer version of cluster-api and thus
	// a new version of controller-runtime.
	sigs.k8s.io/cluster-api-provider-cloudstack => ./internal/thirdparty/capc
)
