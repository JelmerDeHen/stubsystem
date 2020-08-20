package cmds
import (
	"fmt"
	"errors"
	"io/ioutil"
	"unsafe"
	"syscall"
)
// The command should only return int
// This can then become exit status
func Insmod(argv []string) (int) {
	if len(argv) == 0 {
		InsmodUsage(errors.New("Need: module names"))
		return 1
	}
	for _, name := range argv {
		err := insmod(name)
		if err != nil {
			InsmodUsage(err)
			return 2
		}
	}
	return 0
}
func InsmodUsage(err error) {
	fmt.Println(err)
	fmt.Println("Usage: insmod <module[ module...]>")
}
// This function should return only an error
// The resolution should also happen in the function itself
//
// Using syscall init_module
//asmlinkage long sys_init_module(void __user *umod, unsigned long len const char __user *uargs);
func insmod (module string) error {
	path := "/modules/" + module + ".ko"
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if len(buf) == 0 {
		return errors.New("Insmod for " + path + " failed because provided file is empty")
	}
	bufptr := uintptr(unsafe.Pointer(&buf[0]))
	// len
	lenptr := uintptr(len(buf))
	// uargs
	uargs := ""
	var uargsptr *byte
	uargsptr, err = syscall.BytePtrFromString(uargs)
	if err != nil {
		return err
	}
	r1, r2, err := syscall.Syscall(syscall.SYS_INIT_MODULE, bufptr, lenptr, uintptr(unsafe.Pointer(uargsptr)))
	fmt.Println(r1, r2)
	if r1 != 0 {
		return err
	}
	if r2 > 0 {
		fmt.Printf("r2: %d == %X %+v\n", r2, r2, err)
	}
	return nil
}

