# cryptfs storage backend

The `cryptfs` backend is an experimental storage backend **PREVIEW**. It hashes secret names and stores the mapping from names to actual file inside an `age` encrypted lookup table. The filesystem backing this storage backend is flexible, but by default uses `gitfs`.

**WARNING**: Do not use unless you want to contribute to the development of this backend!
