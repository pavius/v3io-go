using Go = import "/go.capnp";
@0x8a6e4e653e2db81e;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("ExtAttrOuter");

using import "/node/common/ExtAttrKey.capnp".ExtAttrKey;
using import "/node/common/ExtAttrValue.capnp".ExtAttrValue;

struct ExtAttr{
    key             @0 : ExtAttrKey;
    value @1 : ExtAttrValue;
}

struct ExtAttrs{
    keyValuePairs   @0 : List(ExtAttr);
}
