# GoAppImage

AppImage manipulation from Go.

Basically a library form of the AppImage struct from [Probonopd's expiremental Go AppImage tools](https://github.com/probonpd/go-appimage) (specifically from the appimaged tool). The only real difference is that things are exported.

Currently the library calls `mksquashfs` for it's functions. Probonopd and I are looking into using a squashfs library, but that doesn't exist yet (at least not what we need it for)

Calls fairly directly to [AppImage C Library](https://github.com/AppImage/AppImageKit). Created for [LinuxPA](https://github.com/CalebQ42/LinuxPA)
