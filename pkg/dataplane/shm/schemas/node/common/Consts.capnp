using Go = import "/go.capnp";
@0x90f379da7c0d6f74;

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("Consts");

const messagesFileAttrSize          :UInt64 = 1;
const messagesFileAttrMode          :UInt64 = 2;
const messagesFileAttrAtime         :UInt64 = 4;
const messagesFileAttrMtime         :UInt64 = 8;
const messagesFileAttrCtime         :UInt64 = 16;
const messagesStreamAttrSeq         :UInt64 = 32;
const messagesStreamAttrShardAttrs  :UInt64 = 64;
const messagesFileAttrUid           :UInt64 = 512;
const messagesFileAttrGid           :UInt64 = 1024;
const messagesFileAttrInodeNumber   :UInt64 = 2048;

const messagesObjectSetAnyMtimeSec  :UInt64 = 0;
const messagesObjectSetAnyMtimeNsec :UInt64 = 0;

const v3ioUIDNobody                 :UInt16 = 65534;
const v3ioGIDNobody                 :UInt16 = 65534;

const messagesAttrDir   :UInt16 = 3679; # = size | mode | atime | mtime | ctime | stream_attrs | gid | uid | inode_number
const messagesAttrFile  :UInt16 = 3647; # = size | mode | atime | mtime | ctime | stream_seq |  gid | uid | inode_number
const messagesAttrAll   :UInt16 = 3711; # = size | mode | atime | mtime | ctime | stream_seq | stream_attrs | gid | uid | inode_number
