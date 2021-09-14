./premake5 gmake
make foundation config=debug_macosx-arm
make https_static config=debug_macosx-arm
make tmbuild config=debug_macosx-arm
