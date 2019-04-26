using Go = import "/go.capnp";
@0xa9483aaf732d3055;

$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

enum InterfaceSubtype {
    none @0;
    s3 @1; # use AWS equivalents or technical names e.g. object, kv, stream?
    dynamo @2;
    kinesis @3;
}
