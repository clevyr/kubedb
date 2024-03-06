package storage

func IsCloud(path string) bool {
	return IsS3(path) || IsGCS(path)
}

func IsCloudDir(path string) bool {
	return IsS3Dir(path) || IsGCSDir(path)
}
