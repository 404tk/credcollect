package wincred

import (
	"bytes"
	"errors"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/404tk/credcollect/common"
	"golang.org/x/sys/windows"
)

// CredentialPersistence describes one of three persistence modes of a credential.
// A detailed description of the available modes can be found on
// Docs: https://docs.microsoft.com/en-us/windows/desktop/api/wincred/ns-wincred-credentialw
type CredentialPersistence uint32

// CredentialAttribute represents an application-specific attribute of a credential.
type CredentialAttribute struct {
	Keyword string
	Value   []byte
}

// Credential is the basic credential structure.
// A credential is identified by its target name.
// The actual credential secret is available in the CredentialBlob field.
type Credential struct {
	TargetName     string
	Comment        string
	LastWritten    time.Time
	CredentialBlob []byte
	Attributes     []CredentialAttribute
	TargetAlias    string
	UserName       string
	Persist        CredentialPersistence
}

func WinCred() []common.WinCred {
	ret := []common.WinCred{}
	creds, err := credManagerList()
	if err != nil {
		return ret
	}
	for i := range creds {
		if len(creds[i].CredentialBlob) < 100 {
			ret = append(ret, common.WinCred{
				TargetName: creds[i].TargetName,
				UserName:   creds[i].UserName,
				PassWord:   string(bytes.Replace(creds[i].CredentialBlob, []byte{0}, []byte{}, -1)),
			})
		}
	}
	return ret
}

// List retrieves all credentials of the Credentials store.
func credManagerList() ([]*Credential, error) {
	creds, err := sysCredEnumerate("", true)
	if err != nil && errors.Is(err, syscall.Errno(1)) {
		// Ignore ERROR_NOT_FOUND and return an empty list instead
		creds = []*Credential{}
		err = nil
	}
	return creds, err
}

var (
	modadvapi32            = windows.NewLazyDLL("advapi32.dll")
	procCredFree      proc = modadvapi32.NewProc("CredFree")
	procCredEnumerate proc = modadvapi32.NewProc("CredEnumerateW")
)

// Interface for syscall.Proc: helps testing
type proc interface {
	Call(a ...uintptr) (r1, r2 uintptr, lastErr error)
}

// https://docs.microsoft.com/en-us/windows/desktop/api/wincred/ns-wincred-credentialw
type sysCREDENTIAL struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     uintptr
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

// https://docs.microsoft.com/en-us/windows/desktop/api/wincred/ns-wincred-credential_attributew
type sysCREDENTIAL_ATTRIBUTE struct {
	Keyword   *uint16
	Flags     uint32
	ValueSize uint32
	Value     uintptr
}

// https://docs.microsoft.com/en-us/windows/desktop/api/wincred/nf-wincred-credenumeratew
func sysCredEnumerate(filter string, all bool) ([]*Credential, error) {
	var count int
	var pcreds uintptr
	var filterPtr *uint16
	if !all {
		filterPtr, _ = windows.UTF16PtrFromString(filter)
	}
	ret, _, err := procCredEnumerate.Call(
		uintptr(unsafe.Pointer(filterPtr)),
		0,
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&pcreds)),
	)
	if ret == 0 {
		return nil, err
	}
	defer procCredFree.Call(pcreds)
	credsSlice := *(*[]*sysCREDENTIAL)(unsafe.Pointer(&reflect.SliceHeader{
		Data: pcreds,
		Len:  count,
		Cap:  count,
	}))
	creds := make([]*Credential, count, count)
	for i, cred := range credsSlice {
		creds[i] = sysToCredential(cred)
	}

	return creds, nil
}

// Convert the given CREDENTIAL struct to a more usable structure
func sysToCredential(cred *sysCREDENTIAL) (result *Credential) {
	if cred == nil {
		return nil
	}
	result = new(Credential)
	result.Comment = windows.UTF16PtrToString(cred.Comment)
	result.TargetName = windows.UTF16PtrToString(cred.TargetName)
	result.TargetAlias = windows.UTF16PtrToString(cred.TargetAlias)
	result.UserName = windows.UTF16PtrToString(cred.UserName)
	result.LastWritten = time.Unix(0, cred.LastWritten.Nanoseconds())
	result.Persist = CredentialPersistence(cred.Persist)
	result.CredentialBlob = goBytes(cred.CredentialBlob, cred.CredentialBlobSize)
	result.Attributes = make([]CredentialAttribute, cred.AttributeCount)
	attrSlice := *(*[]sysCREDENTIAL_ATTRIBUTE)(unsafe.Pointer(&reflect.SliceHeader{
		Data: cred.Attributes,
		Len:  int(cred.AttributeCount),
		Cap:  int(cred.AttributeCount),
	}))
	for i, attr := range attrSlice {
		resultAttr := &result.Attributes[i]
		resultAttr.Keyword = windows.UTF16PtrToString(attr.Keyword)
		resultAttr.Value = goBytes(attr.Value, attr.ValueSize)
	}
	return result
}

// goBytes copies the given C byte array to a Go byte array (see `C.GoBytes`).
// This function avoids having cgo as dependency.
func goBytes(src uintptr, len uint32) []byte {
	if src == uintptr(0) {
		return []byte{}
	}
	rv := make([]byte, len)
	copy(rv, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: src,
		Len:  int(len),
		Cap:  int(len),
	})))
	return rv
}
