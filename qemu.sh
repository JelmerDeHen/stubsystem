#!/bin/sh
# Testing with Qemu
function getmodules () {
	MODULES="/lib/modules/$(uname -r)"
	[ ! -d "${MODULES}" ] && return 1
	mkdir -pv modules
	local FOUND=0
	while read -r MODULE SPAM; do
		printf 'Module: %s\n' "${MODULE}"
		find "${MODULES}" -name ''"${MODULE}"'.ko.xz' | while read -r KO; do
			printf '%s at %s\n' "${MODULE}" "${KO}"
			local OUT="modules/${KO##*/}"
			if [ "${KO: -3}" = ".xz" ]; then
				OUT="${OUT%.xz}"
				echo "${OUT}"
				xzcat "${KO}" > "${OUT}"
			else
				cp -v "${OUT}" modules
			fi
			FOUND=1
		done
	done </proc/modules
}

function getmodules2 () {
	MODULES=/lib/modules/$(uname -r)
	echo $MODULES
}

main () {
	case "${1}" in
	x86_64)
		#echo A
		#go build .
		#[ $? -ne 0 ] && exit
		#getmodules
		#find init modules -print0 | cpio --null --create --verbose --format=newc | gzip --best > initramfs.img
		#find "/lib/modules/$(uname -r)" init -print0 | cpio --null --create --verbose --format=newc | gzip --best > initramfs.img
		#[ $? -ne 0 ] && exit
		#shift 1
		#qemu-system-x86_64 -enable-kvm -kernel /boot/vmlinuz-linux -initrd ./initramfs.img # -nographic -append "console=ttyAMA0,115200 console=tty highres=off console=ttyS0" $@
	       #	-hda hdd.raw $@
		;;
	x86_64_misc)
		shift 1
		go build
		( cd mkinitcpio && go build && ./mkinitcpio; )

		qemu-system-x86_64 -enable-kvm -kernel /boot/vmlinuz-linux -initrd /tmp/initramfs.img.gz -m 4G -nographic -append "console=ttyAMA0,115200 console=tty highres=off console=ttyS0" $@
		;;
	arm64)
		#GOOS=linux GOARCH=arm64 go build .
		#elif [ "${1}" = "arm64" ]; then
		#-machine virt -cpu cortex-a57 -machine type=virt -nographic -smp 1 -m 2048 -kernel aarch64-linux-3.15rc2-buildroot.img --append "console=ttyAMA0" -fsdev local,id=r,path=/home/alex/lsrc/qemu/rootfs/trusty-core,security_model=none -device virtio-9p-device,fsdev=r,mount_tag=r
		#GOOS=linux GOARCH=arm64 go build .
		#[ $? -ne 0 ] && exit
		#qemu-system-aarch64 -M virt -cpu cortex-a57 -kernel ./vmlinuz-aarch64 -initrd ./initramfs.img -nographic -smp 1 -m 2048 -append "ttyAMA0,115200n8"
		;;
	*)
		echo C
		;;
	esac

#	if [ "${1}" = "x86_64" ]; then
#		go build .
#		[ $? -ne 0 ] && exit
#		#getmodules
#		find init modules -print0 | cpio --null --create --verbose --format=newc | gzip --best > initramfs.img
#		#find "/lib/modules/$(uname -r)" init -print0 | cpio --null --create --verbose --format=newc | gzip --best > initramfs.img
#		[ $? -ne 0 ] && exit
#		shift 1
#		qemu-system-x86_64 -enable-kvm -kernel /boot/vmlinuz-linux -initrd ./initramfs.img # -nographic -append "console=ttyAMA0,115200 console=tty highres=off console=ttyS0" $@
#	       #	-hda hdd.raw $@
#	elif [ "${1}" = "arm64" ]; then
#		#-machine virt -cpu cortex-a57 -machine type=virt -nographic -smp 1 -m 2048 -kernel aarch64-linux-3.15rc2-buildroot.img --append "console=ttyAMA0" -fsdev local,id=r,path=/home/alex/lsrc/qemu/rootfs/trusty-core,security_model=none -device virtio-9p-device,fsdev=r,mount_tag=r
#		[ $? -ne 0 ] && exit
#		qemu-system-aarch64 -M virt -cpu cortex-a57 -kernel ./vmlinuz-aarch64 -initrd ./initramfs.img -nographic -smp 1 -m 2048 -append "ttyAMA0,115200n8"
#	else
#		printf '%s <x86_64|arm64>\n' ${FUNCNAME[0]}
#	fi
}
main $@
