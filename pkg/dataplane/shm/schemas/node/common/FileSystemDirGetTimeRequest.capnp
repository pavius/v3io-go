using Go = import "/go.capnp";
@0xb5728dc1eb32d04c;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

struct FileSystemDirGetTimeRequest {
    inodes @0 :List(UInt32);
}

