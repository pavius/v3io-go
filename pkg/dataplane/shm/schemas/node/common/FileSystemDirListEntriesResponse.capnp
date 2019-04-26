using Go = import "/go.capnp";
@0x9381c80d4f5af7de;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

using import "/node/common/DirEntry.capnp".DirEntry;

struct FileSystemDirListEntriesResponse {
    dirlist                 @0 :List(DirEntry);
    numRemainingEntries     @1 :UInt64;
}
