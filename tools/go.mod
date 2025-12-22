module github.com/FerretDB/FerretDB/v2/tools

go 1.25

toolchain go1.25.5

// replace github.com/FerretDB/wire => ../../wire

tool (
	github.com/FerretDB/FerretDB/v2/tools/checkcomments
	github.com/FerretDB/FerretDB/v2/tools/checkdocs
	github.com/FerretDB/FerretDB/v2/tools/checkswitch
	github.com/FerretDB/FerretDB/v2/tools/definedockertag
	github.com/FerretDB/FerretDB/v2/tools/generatechangelog
	github.com/OpenDocDB/cts/opendocdb-cts
	github.com/go-task/task/v3/cmd/task
	github.com/goreleaser/nfpm/v2/cmd/nfpm
	github.com/jstemmer/go-junit-report/v2
	github.com/kisielk/godepgraph
	github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	github.com/quasilyte/go-consistent
	golang.org/x/perf/cmd/benchstat
	golang.org/x/pkgsite/cmd/pkgsite
	golang.org/x/tools/cmd/deadcode
	golang.org/x/tools/cmd/goimports
	golang.org/x/tools/cmd/stringer
	golang.org/x/vuln/cmd/govulncheck
	mvdan.cc/gofumpt
)

require (
	github.com/FerretDB/gh v0.2.0
	github.com/google/go-github/v70 v70.0.1-0.20250402125210-3a3f51bc7c5d
	github.com/rogpeppe/go-internal v1.14.1
	github.com/sethvargo/go-githubactions v1.3.2
	github.com/stretchr/testify v1.11.1
	golang.org/x/tools v0.40.0
)

require (
	cloud.google.com/go v0.110.2 // indirect
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	cloud.google.com/go/storage v1.29.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/FerretDB/wire v0.1.0 // indirect
	github.com/Ladicle/tabwriter v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/OpenDocDB/cts/opendocdb-cts v0.2.3 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794 // indirect
	github.com/alecthomas/chroma/v2 v2.20.0 // indirect
	github.com/alecthomas/kong v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.15 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.68 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.80.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.20 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/caarlos0/go-version v0.2.2 // indirect
	github.com/cavaliergopher/cpio v1.0.1 // indirect
	github.com/chainguard-dev/git-urls v1.0.2 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dominikbraun/graph v0.23.0 // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elliotchance/orderedmap/v3 v3.1.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/getkin/kin-openapi v0.133.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.16.1 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-task/task/v3 v3.46.3 // indirect
	github.com/go-task/template v0.2.0 // indirect
	github.com/go-toolsmith/astcast v1.0.0 // indirect
	github.com/go-toolsmith/astequal v1.0.0 // indirect
	github.com/go-toolsmith/astinfo v0.0.0-20180906194353-9809ff7efb21 // indirect
	github.com/go-toolsmith/pkgload v1.0.0 // indirect
	github.com/go-toolsmith/typep v1.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/licensecheck v0.3.1 // indirect
	github.com/google/rpmpack v0.7.1 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/safehtml v0.0.3-0.20211026203422-d6f0e11a5516 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/goreleaser/chglog v0.7.3 // indirect
	github.com/goreleaser/fileglob v1.4.0 // indirect
	github.com/goreleaser/nfpm/v2 v2.44.0 // indirect
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.65 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.8.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jstemmer/go-junit-report/v2 v2.1.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/kisielk/godepgraph v1.0.0 // indirect
	github.com/kisielk/gotool v1.0.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/muesli/mango v0.2.0 // indirect
	github.com/muesli/mango-cobra v1.3.0 // indirect
	github.com/muesli/mango-pflag v0.1.0 // indirect
	github.com/muesli/roff v0.1.0 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.5.1 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/puzpuzpuz/xsync/v4 v4.2.0 // indirect
	github.com/quasilyte/go-consistent v0.6.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sajari/fuzzy v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/speakeasy-api/jsonpath v0.6.0 // indirect
	github.com/speakeasy-api/openapi-overlay v0.10.2 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/u-root/u-root v0.15.1-0.20251014130006-62f7144b33da // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/woodsbury/decimal128 v1.3.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	gitlab.com/digitalxero/go-conventional-commit v1.0.7 // indirect
	go.mongodb.org/mongo-driver v1.17.3 // indirect
	go.mongodb.org/mongo-driver/v2 v2.2.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/perf v0.0.0-20251208221838-04cf7a2dca90 // indirect
	golang.org/x/pkgsite v0.0.0-20251221001151-a7c442171859 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/telemetry v0.0.0-20251203150158-8fff8a5912fc // indirect
	golang.org/x/term v0.38.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools/go/packages/packagestest v0.1.1-deprecated // indirect
	golang.org/x/vuln v1.1.4 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.126.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
	mvdan.cc/sh/moreinterp v0.0.0-20251109230715-65adef8e2c5b // indirect
	mvdan.cc/sh/v3 v3.12.0 // indirect
	rsc.io/markdown v0.0.0-20231214224604-88bb533a6020 // indirect
)
