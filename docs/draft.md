ManifestEnvelope
Manifest

loadPartition :: string -> Partition

Partition {
    path
    manifest: Manifest
}

hash :: Partition -> ManifestChange[]

apply :: Partition -> ManifestChange[] -> Partition

persist :: Partition -> IO ()
