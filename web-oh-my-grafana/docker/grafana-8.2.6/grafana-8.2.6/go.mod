module github.com/grafana/grafana

go 1.17

// Override xorm's outdated go-mssqldb dependency, since we can't upgrade to current xorm (due to breaking changes).
// We need a more current go-mssqldb so we get rid of a version of apache/thrift with vulnerabilities.
// Also, use our fork with fixes for unimplemented methods (required for Go 1.16).
replace github.com/denisenkom/go-mssqldb => github.com/grafana/go-mssqldb v0.0.0-20210326084033-d0ce3c521036

// Override k8s.io/client-go outdated dependency, which is an indirect dependency of grafana/loki.
// It's also present on grafana/loki's go.mod so we'll need till it gets updated.
replace k8s.io/client-go => k8s.io/client-go v0.22.1

// For now we are using a Grafana fork of alertmanager until the ability to bypass label name/key validation
// is merged into upstream (2021-10-01).
replace github.com/prometheus/alertmanager => github.com/grafana/alertmanager v0.21.1-0.20210929114340-613e5cc2a4ee

require (
	cloud.google.com/go/storage v1.14.0
	cuelang.org/go v0.3.2
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.19.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.10.0
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/semver v1.5.0
	github.com/VividCortex/mysqlerr v0.0.0-20170204212430-6c6b55f8796f
	github.com/aws/aws-sdk-go v1.40.11
	github.com/beevik/etree v1.1.0
	github.com/benbjohnson/clock v1.1.0
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/centrifugal/centrifuge v0.18.4
	github.com/cortexproject/cortex v1.8.2-0.20210428155238-d382e1d80eaf
	github.com/crewjam/saml v0.4.6-0.20201227203850-bca570abb2ce
	github.com/davecgh/go-spew v1.1.1
	github.com/denisenkom/go-mssqldb v0.0.0-20200910202707-1e08a3fab204
	github.com/dop251/goja v0.0.0-20210804101310-32956a348b49
	github.com/fatih/color v1.10.0
	github.com/gchaincl/sqlhooks v1.3.0
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-kit/kit v0.11.0
	github.com/go-macaron/binding v0.0.0-20190806013118-0b4f37bab25b
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack v1.8.0
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug v1.9.0
	github.com/grafana/grafana-aws-sdk v0.7.0
	github.com/grafana/grafana-plugin-sdk-go v0.114.0
	github.com/grafana/loki v1.6.2-0.20210520072447-15d417efe103
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/go-plugin v1.4.2
	github.com/hashicorp/go-version v1.3.0
	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec
	github.com/influxdata/influxdb-client-go/v2 v2.2.3
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.11
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/laher/mergefs v0.1.1
	github.com/lib/pq v1.10.0
	github.com/linkedin/goavro/v2 v2.10.0
	github.com/m3db/prometheus_remote_client_golang v0.4.4
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f
	github.com/opentracing/opentracing-go v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/browser v0.0.0-20210904010418-6d279e18f982 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.23.0-rc.0.0.20210906104939-8da517524a87
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/prometheus v1.8.2-0.20210621150501-ff58416a0b02
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/robfig/cron/v3 v3.0.1
	github.com/russellhaering/goxmldsig v1.1.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/spyzhov/ajson v0.4.2
	github.com/stretchr/testify v1.7.0
	github.com/teris-io/shortid v0.0.0-20171029131806-771a37caa5cf
	github.com/timberio/go-datemath v0.1.1-0.20200323150745-74ddef604fff
	github.com/ua-parser/uap-go v0.0.0-20190826212731-daf92ba38329
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/unknwon/com v1.0.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/weaveworks/common v0.0.0-20210419092856-009d1eebd624
	github.com/xorcare/pointer v1.1.0
	github.com/yudai/gojsondiff v1.0.0
	go.opentelemetry.io/collector v0.31.0
	go.opentelemetry.io/collector/model v0.31.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/exp v0.0.0-20210220032938-85be41e4509f // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	golang.org/x/tools v0.1.5
	gonum.org/v1/gonum v0.9.1
	google.golang.org/api v0.54.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.62.0
	gopkg.in/ldap.v3 v3.1.0
	gopkg.in/macaron.v1 v1.4.0
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	xorm.io/builder v0.3.6
	xorm.io/core v0.7.3
	xorm.io/xorm v0.8.2
)

require (
	cloud.google.com/go v0.90.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v0.7.0 // indirect
	github.com/FZambia/eagle v0.0.1 // indirect
	github.com/FZambia/sentinel v1.1.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20210223225224-5bea62493d91 // indirect
	github.com/armon/go-metrics v0.3.8 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/centrifugal/protocol v0.7.3 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/cockroachdb/apd/v2 v2.0.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/deepmap/oapi-codegen v1.3.13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.4.1-0.20201116162257-a2a8dda75c91 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/emicklei/proto v1.6.15 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-kit/log v0.1.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/loads v0.20.2 // indirect
	github.com/go-openapi/runtime v0.19.29 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/status v1.0.3 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/flatbuffers v1.12.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.1-0.20191002090509-6af20e3a5340 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/memberlist v0.2.4 // indirect
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/igm/sockjs-go/v3 v3.0.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/cpuid/v2 v2.0.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.0.0-20201213122252-bcd7e1b9601e // indirect
	github.com/mattetti/filebuffer v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.1.42 // indirect
	github.com/mitchellh/go-testing-interface v1.14.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mna/redisc v1.2.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mpvl/unique v0.0.0-20150818121801-cbe035fff7de // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/exporter-toolkit v0.6.1 // indirect
	github.com/prometheus/node_exporter v1.0.0-rc.0.0.20200428091818-01054558c289 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/rs/cors v1.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/segmentio/encoding v0.2.19 // indirect
	github.com/sercand/kuberesolver v2.4.0+incompatible // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/weaveworks/promrus v1.2.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	go.mongodb.org/mongo-driver v1.5.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.10 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210805201207-89edb61ffb67 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
)

replace github.com/apache/thrift => github.com/apache/thrift v0.14.1

replace gopkg.in/macaron.v1 => ./pkg/macaron

replace github.com/go-macaron/binding => ./pkg/macaron/binding

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.10.2
