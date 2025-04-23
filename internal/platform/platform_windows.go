package platform

import (
	"errors"
	"strconv"
	"unsafe"

	"github.com/danieljoos/wincred"
	"github.com/vzhd1701/evertoken/internal/myerrors"
	"golang.org/x/sys/windows"
)

func GetDiskSerial() string {
	mainDrive := "C"

	lpVolumeNameBuffer := make([]uint16, 256)
	lpVolumeSerialNumber := uint32(0)
	lpMaximumComponentLength := uint32(0)
	lpFileSystemFlags := uint32(0)
	lpFileSystemNameBuffer := make([]uint16, 256)
	volpath, _ := windows.UTF16PtrFromString(mainDrive + ":/")

	lpVolumeNameBufferPtr := (*uint16)(unsafe.Pointer(&lpVolumeNameBuffer))
	lpFileSystemNameBufferPtr := (*uint16)(unsafe.Pointer(&lpFileSystemNameBuffer))

	err := windows.GetVolumeInformation(
		volpath,
		lpVolumeNameBufferPtr,
		uint32(len(lpVolumeNameBuffer)),
		&lpVolumeSerialNumber,
		&lpMaximumComponentLength,
		&lpFileSystemFlags,
		lpFileSystemNameBufferPtr,
		uint32(len(lpFileSystemNameBuffer)))
	myerrors.PanicFail(err)

	return strconv.Itoa(int(lpVolumeSerialNumber))
}

func GetSecureStorageData(service string, accountID string) []byte {
	serviceURI := service + "/" + accountID

	cred, err := wincred.GetGenericCredential(serviceURI)
	if err != nil {
		switch err.Error() {
		case "Element not found.":
			err = errors.New("entry not found in secure storage")
		default:
			myerrors.PanicFail(err)
		}
		myerrors.ExpectedFail(err)
	}

	return cred.CredentialBlob
}
