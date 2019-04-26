using Go = import "/go.capnp";
@0xb01a9d2d24c0825b;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("ExtAttrKeyOuter");

struct ExtAttrKey{
    name            @0 :Text;
    schemaId        @1 :UInt8;
}

struct ExtAttrKeys{
    keys @0 :List(ExtAttrKey);
}

