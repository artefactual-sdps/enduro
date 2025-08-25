module github.com/artefactual-sdps/enduro

go 1.24.6

require (
	ariga.io/atlas v0.32.0
	ariga.io/sqlcomment v0.1.0
	buf.build/gen/go/artefactual/a3m/grpc/go v1.3.0-20230508184533-2e9432075630.2
	buf.build/gen/go/artefactual/a3m/protocolbuffers/go v1.31.0-20230508184533-2e9432075630.2
	chainguard.dev/go-oidctest v0.3.1
	entgo.io/ent v0.14.4
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.0
	github.com/XSAM/otelsql v0.29.0
	github.com/alicebob/miniredis/v2 v2.32.1
	github.com/artefactual-sdps/temporal-activities v0.0.0-20250609161607-98135eaabe3f
	github.com/coreos/go-oidc/v3 v3.10.0
	github.com/cyphar/filepath-securejoin v0.2.4
	github.com/dolmen-go/contextio v1.0.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/gliderlabs/ssh v0.3.6
	github.com/go-jose/go-jose/v3 v3.0.4
	github.com/go-logr/logr v1.4.2
	github.com/go-sql-driver/mysql v1.9.1
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/jmoiron/sqlx v1.3.5
	github.com/jonboulle/clockwork v0.4.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mholt/archiver/v4 v4.0.0-alpha.8
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nyudlts/go-bagit v0.3.0-alpha.0.20240515212815-8dab411c23af
	github.com/oklog/run v1.1.0
	github.com/otiai10/copy v1.14.0
	github.com/pkg/sftp v1.13.6
	github.com/prometheus/client_golang v1.19.0
	github.com/radovskyb/watcher v1.0.7
	github.com/redis/go-redis/extra/redisotel/v9 v9.0.5
	github.com/redis/go-redis/v9 v9.10.0
	github.com/rukavina/sftpblob v0.0.0-20201030103652-e8e9601e6511
	github.com/spf13/afero v1.11.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.10.0
	go.artefactual.dev/amclient v0.4.1-0.20240705155055-0c5abef5207c
	go.artefactual.dev/tools v0.21.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.54.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.29.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.temporal.io/api v1.44.1
	go.temporal.io/sdk v1.33.1
	go.temporal.io/sdk/contrib/opentelemetry v0.6.0
	go.uber.org/mock v0.4.0
	goa.design/goa/v3 v3.15.2
	goa.design/plugins/v3 v3.15.2
	gocloud.dev v0.41.0
	golang.org/x/crypto v0.37.0
	golang.org/x/exp v0.0.0-20231219180239-dc181d75b848
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
	gotest.tools/v3 v3.5.2
)

require (
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/AnatolyRugalev/goregen v0.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/STARRY-S/zip v0.2.1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/alicebob/gopher-json v0.0.0-20230218143504-906a9b012302 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/artefactual-labs/bagit-gython v0.2.0 // indirect
	github.com/artefactual-labs/bine v0.20.0 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.12 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.65 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.69 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.78.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.17 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.0 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimfeld/httppath v0.0.0-20170720192232-ee938bf73598 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/renameio/v2 v2.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/safeopen v0.0.0-20240125081138-66b54d5181c6 // indirect
	github.com/google/wire v0.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl/v2 v2.19.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kluctl/go-embed-python v0.0.0-3.12.3-20240415-1 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/manveru/faker v0.0.0-20171103152722-9fbc68a78c4d // indirect
	github.com/mholt/archives v0.1.1 // indirect
	github.com/minio/minlz v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/nexus-rpc/sdk-go v0.3.0 // indirect
	github.com/nwaples/rardecode/v2 v2.1.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.0.5 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sorairolake/lzip-go v0.3.5 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tailscale/hujson v0.0.0-20250226034555-ec1d1c113d33 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.228.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool github.com/artefactual-labs/bine
