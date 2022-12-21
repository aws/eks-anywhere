package tar

type GzipPackager struct{}

func NewGzipPackager() GzipPackager {
	return GzipPackager{}
}

func (GzipPackager) Package(sourceFolder, dstFile string) error {
	return GzipTarFolder(sourceFolder, dstFile)
}

func (GzipPackager) UnPackage(orgFile, dstFolder string) error {
	return UnGzipTarFile(orgFile, dstFolder)
}
