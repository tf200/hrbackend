package pdf

import pkgbucket "hrbackend/pkg/bucket"

type pdfService struct {
	bucketClient *pkgbucket.ObjectStorageClient
}

func NewPdfService(bucketClient *pkgbucket.ObjectStorageClient) *pdfService {
	return &pdfService{
		bucketClient: bucketClient,
	}
}
