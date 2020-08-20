package main

import (
	"fmt"
	"io/ioutil"
	"time"
	"syscall"
	"os"
	"os/exec"
	"bufio"
	"strings"
	//"errors"
	//"path/filepath"
	"strconv"
	//"unsafe"
	"github.com/JelmerDeHen/stubsystem/idlebox/cmds"
)
const (
	LINUX_REBOOT_MAGIC1         uintptr = 0xfee1dead
	LINUX_REBOOT_MAGIC2         uintptr = 0x28121969
	LINUX_REBOOT_MAGIC2A        uintptr = 0x5121996
	LINUX_REBOOT_MAGIC2B        uintptr = 0x16041998
	LINUX_REBOOT_MAGIC2C        uintptr = 0x20112000
	LINUX_REBOOT_CMD_CAD_OFF    uintptr = 0
	LINUX_REBOOT_CMD_CAD_ON     uintptr = 0x89abcdef
	LINUX_REBOOT_CMD_HALT       uintptr = 0xcdef0123
	LINUX_REBOOT_CMD_KEXEC      uintptr = 0x45584543
	LINUX_REBOOT_CMD_POWER_OFF  uintptr = 0x4321fedc
	LINUX_REBOOT_CMD_RESTART    uintptr = 0x1234567
	LINUX_REBOOT_CMD_RESTART2   uintptr = 0xa1b2c3d4
	LINUX_REBOOT_CMD_SW_SUSPEND uintptr = 0xd000fce1
)
func ls() {
	files, err := ioutil.ReadDir("/")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range files {
		fmt.Println(f.Name())
		files2, err := ioutil.ReadDir(f.Name())
		if err != nil {
			continue
		}
		for _, f2 := range files2 {
			fmt.Println("	", f2.Name())
			//time.Sleep(time.Second/10)
		}
	}
}

const (
	MS_RDONLY		uintptr = 1	// Mount read-only
	MS_NOSUID		uintptr = 2	// Ignore suid and sgid bits
	MS_NODEV		uintptr = 4	// Disallow access to device special files
	MS_NOEXEC		uintptr = 8	// Disallow program execution
	MS_SYNCHRONOUS		uintptr = 16	// Writes are synced at once
	MS_REMOUNT		uintptr = 32	// Alter flags of a mounted FS
	MS_MANDLOCK		uintptr = 64	// Allow mandatory locks on an FS
	MS_DIRSYNC		uintptr = 128	// Directory modifications are synchronous

	MS_NOATIME		uintptr = 1024	// Do not update access times
	MS_NODIRATIME		uintptr = 2048	// Do not update directory access times
	MS_BIND			uintptr = 4096
	MS_MOVE			uintptr	= 8192
	MS_REC			uintptr	= 16384
	MS_VERBOSE		uintptr = 32768	// War is peace. Verbosity is silence. (MS_VERBOSE is deprecated)
	MS_SILENT		uintptr = 32768
	MS_POSIXACL		uintptr = (1<<16)	// VFS does not apply the umask
	MS_UNBINDABLE		uintptr = (1<<17)	// change to unbindable
	MS_PRIVATE		uintptr = (1<<18)	// change to private
	MS_SLAVE		uintptr = (1<<19)	// change to slave
	MS_SHARED		uintptr = (1<<20)	// change to shared
	MS_RELATIME		uintptr = (1<<21)	// Update atime relative to mtime/ctime.
	MS_KERNMOUNT		uintptr = (1<<22)	// this is a kern_mount call
	MS_I_VERSION		uintptr = (1<<23)	// Update inode I_version field
	MS_STRICTATIME		uintptr = (1<<24)	// Always perform atime updates
	MS_LAZYTIME		uintptr = (1<<25)	// Update the on-disk [acm]times lazily
	// These sb flags are internal to the kernel
	MS_SUBMOUNT		uintptr = (1<<26)
	MS_NOREMOTELOCK		uintptr = (1<<27)
	MS_NOSEC		uintptr = (1<<28)
	MS_BORN			uintptr = (1<<29)
	MS_ACTIVE		uintptr = (1<<30)
	MS_NOUSER		uintptr = (1<<31)
	// Superblock flags that can be altered by MS_REMOUNT
	MS_RMT_MASK		uintptr = (MS_RDONLY|MS_SYNCHRONOUS|MS_MANDLOCK|MS_I_VERSION|MS_LAZYTIME)
	// Old magic mount flag and mask
	MS_MGC_VAL		uintptr = 0xC0ED0000
	MS_MGC_MSK		uintptr = 0xffff0000


)
// mount(2) needs CAP_SYS_ADMIN
func mount () error {
	os.Mkdir("/proc", 0666)
	os.Mkdir("/sys", 0666)
	os.Mkdir("/dev", 0766)
	os.Mkdir("/run", 0766)
	err := syscall.Mount("", "/proc", "proc", MS_NOSUID|MS_NOEXEC|MS_NODEV, "")
	if err != nil {
		return err
	}
	err = syscall.Mount("", "/sys", "sysfs", MS_NOSUID|MS_NOEXEC|MS_NODEV, "")
	if err != nil {
		return err
	}
	err = syscall.Mount("", "/dev", "devtmpfs", MS_NOSUID, "")
	if err != nil {
		return err
	}
	err = syscall.Mount("", "/run", "tmpfs", MS_NOSUID|MS_NODEV, "")

	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		err = syscall.Mount("", "/sys/firmware/efi/efivars", "efivarfs", MS_NOSUID|MS_NODEV|MS_NOEXEC, "")
		if err != nil {
			return err
		}
	}
	return nil
}
type Console struct {
	Pwd string
}
func (c *Console) Spawn() {
	for {
		fmt.Print("> ")
		r := bufio.NewReader(os.Stdin)
		cmd, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		if cmd == "exit\n" {
			break
		}
		c.Exec(cmd)
	}
}

func (c *Console) Ls(dirs []string) {
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	for _, dir := range dirs {
//		files := make([]os.FileInfo, 0)
		dot, _ := os.Lstat(".")
		dotdot, _ := os.Lstat("..")
		files := []os.FileInfo{dot, dotdot}
		f, err := os.Open(dir)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer f.Close()
		fileInfo, err := f.Readdir(-1)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, file := range fileInfo {
			files = append(files, file)
		}

		//files, err := filepath.Glob(dir)
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
		//fmt.Println(files)
		fmt.Println("Mode\t\tNlink\tUid\tGid\tSize\tYYYY\tM\tD\tT\tName")
		for _, fi := range files {
//			fmt.Println(fi)
			sysinfo := fi.Sys().(*syscall.Stat_t)
			sec, nsec := sysinfo.Ctim.Unix()
			ts := time.Unix(sec, nsec)
			y,m,d := ts.Date()
			fmt.Printf("%s\t%d\t%d\t%d\t%d\t%d\t%s\t%02d\t%02d:%02d\t%s\n", fi.Mode().String(), sysinfo.Nlink, sysinfo.Uid, sysinfo.Gid, fi.Size(), y, m.String()[:3], d, ts.Hour(), ts.Minute(), fi.Name())
			//fmt.Printf("%v\n", sysinfo)
		}
	}
}


/*
func (c *Console) Insmod(modules []string) {
	if len(modules) == 0 {
		fmt.Println("Need: module names")
		return
	}
	for _, module := range modules {
		path := "/modules/" + module + ".ko"
		err := Insmod(path)
		if err != nil {
			fmt.Println(err)
		}
	}
}
*/
//asmlinkage long sys_init_module(void __user *umod, unsigned long len, const char __user *uargs);
/*
func Insmod(path string) (error) {
	// buf
	buf, err := ioutil.ReadFile(path)
	fmt.Println("HERE", path, buf, err)
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
*/
func (c *Console) Cat(argv []string) {
	for _, file := range argv {
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(string(dat))
	}
}
func (c *Console) Cd(argv []string) {
	if len(argv) == 1 {
		os.Chdir(argv[0])
	}
}
func (c *Console) Mount(argv []string) {
	if len(argv) != 5 {
		fmt.Println("mount <source> <target> <fstype> <flags> <data>")
		return
	}
	source := argv[0]
	target := argv[1]
	fstype := argv[2]
	flags, err := strconv.Atoi(argv[3])
	if err != nil {
		fmt.Println("flags needs to be digit:", err)
		return
	}
	flagsptr := uintptr(flags)
	data := argv[4]
	fmt.Printf("source=%s; target=%s; fstype=%s; flags=%d; data=%s\n")
	err = syscall.Mount(source, target, fstype, flagsptr, data)
	if err != nil{
		fmt.Println(err)
		return
	}
}
func (c *Console) Env(argv []string) {
	fmt.Println(os.Environ())
}
func (c *Console) Mkdir(argv []string) {
	for _, dir := range argv {
		err := os.Mkdir(dir, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func (c *Console) Exec(cmd string) {
	cmd = strings.TrimSpace(cmd)
	split := strings.Split(cmd, " ")
	if len(split) == 0 {
		return
	}
	cmd = split[0]
	argv := split[1:]
	var code int
	_=code
	switch split[0] {
	case "ls":
		c.Ls(argv)
	case "cd":
		c.Cd(argv)
	case "cat":
		c.Cat(argv)
	case "mount":
		c.Mount(argv)
	case "env":
		c.Env(argv)
	case "mkdir":
		c.Mkdir(argv)
	case "insmod":
		//c.Insmod(argv)
		cmdOut, err := exec.Command("/usr/bin/insmod", argv...).Output()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(cmdOut)
		}
	default:
		fmt.Println("Not implemented:", cmd)
	}
	//if strings.TrimSpace(slice[0])
}

func forEachFile(dir string, cb func (path string) ()) (error) {
	fh, err := os.Open(dir)
	if err != nil {
		return err
	}
	fis, err := fh.Readdir(-1)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		cb(dir + "/" + fi.Name())
	}
	return nil
}
func printk(msg string) {
	// Maybe have to mknod /dev/console?
	err := ioutil.WriteFile("/dev/console", []byte(msg), 0644)
	if err!=nil {
		return
	}
}

func rdlogger(msg string) (err error) {
	// rd.log=
	err = ioutil.WriteFile("/dev/console", []byte(msg), 0644)
	if err != nil {
		return
	}
	// rd.log=kmsg
	err = ioutil.WriteFile("/dev/kmsg", []byte(msg), 0644)
	if err != nil {
		return
	}
	// 
	err = ioutil.WriteFile("/dev/kmsg", []byte(msg), 0644)
	if err != nil {
		return
	}
	return
}

func main() {
	//	http.Get("http://google.com")
	fmt.Println("Hello world")
	mount()
	rdlogger("Hi")
	printk("~> Hello World!?\n")

	// Mount all modules skipping a few problematic ones

	/*
	includes := []string{"hid", "usbhid", "atkbd"}
	forEachFile("/modules", func ( path string) () {
		ok := false
		for _, filter := range includes {
			if strings.Contains(path, filter) {
				fmt.Println("ADD", path)
				ok = true
			}
		}
		if ok == true {
			fmt.Println("+", path)
			Insmod(path)
		} else {
			fmt.Println("-", path)
		}
	})
	*/
	x := cmds.Insmod([]string{"/modules/hid.ko"})
	fmt.Println(x)

	fmt.Println("AAA")
	c := &Console{}
	c.Spawn()
	//c.Ls([]string{"/"})
	fmt.Println("AAA")
	for i := 3; i > 0; i-- {
		rdlogger(fmt.Sprintln(i))
		time.Sleep(time.Second)
	}
	syscall.Syscall(syscall.SYS_REBOOT, LINUX_REBOOT_MAGIC1, LINUX_REBOOT_MAGIC2, LINUX_REBOOT_CMD_POWER_OFF)

}
