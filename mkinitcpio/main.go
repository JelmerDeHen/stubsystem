package main

import (
	"github.com/cavaliercoder/go-cpio"
	"github.com/ulikunitz/xz"
	"fmt"
	"os"
	"strings"
//	"encoding/binary"
	"strconv"
	"path/filepath"
	"io/ioutil"
	"bytes"
	"compress/gzip"
	"flag"
)

// Async now
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
		cb(fi.Name())
	}
	return nil
}
func CatInt (fn string) (int, error) {
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	bufsize, n := 9, 9
	buf := make([]byte, 9)
	for i:=0; i<10 && n == bufsize; i++ {
		n, err = f.Read(buf)
		if err != nil  {
			return 0, err
		}
	}
	sbuf := string(buf)
	sbuf = strings.TrimRight(sbuf, "\x00\n")
	i, err := strconv.Atoi(sbuf)
	if err != nil {
		return 0, err
	}
	return i, nil
}

type ProcModuleT struct {
	Name string
	Size int
	Refcnt int
	Usedby []string
}
// Based on `strace lsmod`
func Lsmod (sysmodule string) (ret []*ProcModuleT) {
	forEachFile(sysmodule, func(moduleName string) {
		module := &ProcModuleT{Name: moduleName}
		fh, err := os.Open(sysmodule + "/" + moduleName + "/holders")
		if err == nil {
			fis, err := fh.Readdir(-1)
			if err == nil {
				for _, fi := range fis {
					module.Usedby = append(module.Usedby, fi.Name())
				}
			}
			fh.Close()
		}
		refcnt, err := CatInt(sysmodule + "/" + moduleName + "/refcnt")
		if err == nil {
			module.Refcnt = refcnt
		}
		coresize, err := CatInt(sysmodule + "/" + moduleName + "/coresize")
		if err == nil {
			module.Size = coresize
		}
		//fmt.Println(module)
		ret = append(ret, module)
	})
	return ret
}
//func ResolveKmod(kmodbasedir, kmname string) {
//}


func Unxz(compressed []byte) ([]byte, error) {
	xzr, err := xz.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(xzr)
	return buf.Bytes(), nil

}
/*
type Initrd struct {

}
func (* Initrd) Add(name string, buf []byte) (error) {
	return nil
}
*/
const (
	CONFIG_INITRD_OUT =	"/boot/init/initramfs.img"
	CONFIG_INIT =		"../entry/entry"
	CONFIG_IDLEBOX =	"../idlebox/idlebox"
)
func main () {
	kver, err := UnameR("/proc/version")
	// modinfo calls this: -k, --set-version=VERSION   Use VERSION instead of `uname -r`
	kmodbaseFlag := flag.String("k", "/lib/modules/" + kver, "Use path for kernel modules")
	initrdOut := flag.String("o", CONFIG_INITRD_OUT, "Write initramfs here")
	flag.Parse()

	f, err := os.OpenFile(*initrdOut, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var initrd * cpio.Writer
	initrd = cpio.NewWriter(f)
	// Add init
	buf, err := ioutil.ReadFile("../entry/entry")
	if err != nil {
		fmt.Println(err)
		return
	}
	hdr := &cpio.Header{Name: "init", Mode: 0700, Size: int64(len(buf))}
	if err := initrd.WriteHeader(hdr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := initrd.Write(buf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Create "modules" dir
	hdr = &cpio.Header{
		Name: "modules",
		Mode: cpio.ModeDir | 0700,
	}
	if err := initrd.WriteHeader(hdr); err != nil {
		fmt.Println("Error creating modules")
		fmt.Println(err)
		os.Exit(2)
	}

	err = AddModulesEx(initrd, *kmodbaseFlag)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Add idlebox
	AddIdlebox(initrd)

	if err := initrd.Close(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now gzip /tmp/initramfs.img
	f.Close()
	// Read entire file and gzip it
	var b bytes.Buffer
	zw :=  gzip.NewWriter(&b)
	data, err := ioutil.ReadFile(*initrdOut)
	if err != nil {
		fmt.Println(err)
		return
	}
	zw.Write(data)
	// Now overwrite the old file with gz
	f, err = os.OpenFile(*initrdOut + ".gz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(b.Bytes())
	fmt.Printf("Ready: %s\n", *initrdOut)
}
func AddModules(kmodbase string, initrd  *cpio.Writer) error {
	if _, err := os.Stat(kmodbase); err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}
	filepath.Walk(kmodbase, func(path string, fi os.FileInfo, err error) error {
		if strings.HasSuffix(fi.Name(), ".ko") || strings.HasSuffix(fi.Name(), ".ko.xz") {
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			// Decompress .xz
			if strings.HasSuffix(path, ".xz") {
				unxz, err := Unxz(buf)
				if err != nil {
					fmt.Println(err)
				}
				buf = unxz
			}
			relpath := strings.TrimSuffix("modules/" + fi.Name(), ".xz")
			hdr := &cpio.Header{Name: relpath, Mode: 0700, Size: int64(len(buf))}
			if err := initrd.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := initrd.Write(buf); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}
func addKObject (initrd *cpio.Writer, path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// Decompress .xz
	if strings.HasSuffix(stat.Name(), ".xz") {
		unxz, err := Unxz(buf)
		if err != nil {
			fmt.Println(err)
		}
		// Update buffer
		buf = unxz
	}
	relpath := strings.TrimSuffix("modules/" + stat.Name(), ".xz")
	hdr := &cpio.Header{Name: relpath, Mode: 0700, Size: int64(len(buf))}

	if err := initrd.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := initrd.Write(buf); err != nil {
		return err
	}
	return nil
}
func parseModulesBin () {
	// /usr/lib/modules/5.6.12-arch1-1/modules.dep.bin
	// https://github.com/torvalds/linux/blob/3243a89dcbd8f5810b72ee0903d349bd000c4c9d/scripts/depmod.sh
}
func testModule () {
	modules, err := getLoadedModules()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, module := range(modules) {
		fmt.Println(module)
		koPath, err := getModulePath("5.6.12-arch1-1", module)
		if err != nil {
			fmt.Println("Err", err)
			return
		}
		fmt.Println(koPath)
	}
	if len(os.Args) == 2 {
		koPath, err := getModulePath("5.6.12-arch1-1", os.Args[1])
		if err != nil {
			fmt.Println("err")
			fmt.Println(err)
			return
		}
		fmt.Println(koPath)
	}
}
// Figures out the currently loaded modules and adds them
func AddModulesEx (initrd  *cpio.Writer, kver string) (error) {
	modules, err := getLoadedModules()
	if err != nil {
		return err
	}
	var kobjectpaths []string
	for _, mod:= range(modules) {
		kobjectpath, err := getModulePath(kver, mod)
		if err != nil {
			return err
		}
		kobjectpaths = append(kobjectpaths, kobjectpath)
	}
	for _, ko := range kobjectpaths {
		err = addKObject(initrd, ko)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddIdlebox(initrd *cpio.Writer) error {
	hdr := &cpio.Header{
		Name: "usr",
		Mode: cpio.ModeDir | 0700,
	}
	initrd.WriteHeader(hdr)
	hdr = &cpio.Header{
		Name: "usr/bin",
		Mode: cpio.ModeDir | 0700,
	}
	initrd.WriteHeader(hdr)


	data, err := ioutil.ReadFile("../idlebox/idlebox")
	if err != nil {
		return err
	}
	hdr = &cpio.Header{
		Name: "usr/bin/idlebox",
		Mode: 0700,
		Size: int64(len(data)),
	}
	initrd.WriteHeader(hdr)
	if _, err := initrd.Write(data); err != nil {
		return err
	}
	/*
	hdr = &cpio.Header{
		Name: "usr/bin/ls",
		Mode: os.ModeSymlink,
		Linkname: "/usr/bin/idlebox",
	}
	initrd.WriteHeader(hdr)
	hdr = &cpio.Header{
		Name: "usr/bin/insmod",
		Mode: cpio.ModeSymlink,
		Linkname: "/usr/bin/idlebox",
	}
	initrd.WriteHeader(hdr)
	hdr = &cpio.Header{
		Name: "usr/bin/cat",
		Mode: cpio.ModeSymlink,
		Linkname: "/usr/bin/idlebox",
	}
	initrd.WriteHeader(hdr)
	*/
	return nil
}
