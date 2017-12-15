from .event import Event
from .reader import Reader
from .writer import Writer
from .proto import BucketHeader

GZIP = BucketHeader.GZIP
LZ4 = BucketHeader.LZ4
UNCOMPRESSED = BucketHeader.NONE
