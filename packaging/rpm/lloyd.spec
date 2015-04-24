Summary:    Line-delimited JSON utils.
Name:       lloyd
Version:    0.2.4
Release:    0
License:    MIT
BuildRoot:  %{_tmppath}/%{name}-build
BuildArch:  x86_64
Group:      System/Base
Vendor:     Leipzig University Library, https://www.ub.uni-leipzig.de
URL:        https://github.com/miku/lloyd

%description

Line-delimited JSON utils.

%prep
# the set up macro unpacks the source bundle and changes in to the represented by
# %{name} which in this case would be my_maintenance_scripts. So your source bundle
# needs to have a top level directory inside called my_maintenance _scripts
# %setup -n %{name}

%build
# this section is empty for this example as we're not actually building anything

%install
# create directories where the files will be located
mkdir -p $RPM_BUILD_ROOT/usr/local/sbin

# put the files in to the relevant directories.
# the argument on -m is the permissions expressed as octal. (See chmod man page for details.)
install -m 755 lloyd-map $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 lloyd-permute $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 lloyd-uniq $RPM_BUILD_ROOT/usr/local/sbin

%post
# the post section is where you can run commands after the rpm is installed.
# insserv /etc/init.d/my_maintenance

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

%files
%defattr(-,root,root)
/usr/local/sbin/lloyd-map
/usr/local/sbin/lloyd-permute
/usr/local/sbin/lloyd-uniq

%changelog
* Thu Apr 23 2015 Martin Czygan
- 0.2.1, first public release

* Thu Apr 23 2015 Martin Czygan
- 0.2.0, remove lloyd-uniq

* Wed Apr 22 2015 Martin Czygan
- 0.1.1, initial release
