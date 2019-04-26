using Go = import "/go.capnp";
@0x883d416d4f7ebb97;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("V3ioShardData");

struct ShardData {
    metadataArraySize           @0 :UInt16;
    metadataArrayGranularity    @1 :UInt32;
    metadataCurrChunkSeqNum     @2 :UInt32;
}