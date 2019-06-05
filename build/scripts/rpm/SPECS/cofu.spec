Name:           %{_product_name}
Version:        %{_product_version}

Release:        1.el%{_rhel_version}
Summary:        Minimum configuration management tool written in Go.
Group:          Development/Tools
License:        MIT
Source0:        %{name}_linux_amd64.zip
Source1:        cofu-agent.toml
Source2:        cofu-agent.sysconfig
Source3:        cofu-agent.service
BuildRoot:      %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

%description
Minimum configuration management tool written in Go.

%prep
%setup -q -c

%install
mkdir -p %{buildroot}/%{_bindir}
cp %{name} %{buildroot}/%{_bindir}

mkdir -p %{buildroot}/%{_sysconfdir}/cofu-agent
cp %{SOURCE1} %{buildroot}/%{_sysconfdir}/cofu-agent/cofu-agent.toml
mkdir -p %{buildroot}/%{_sysconfdir}/cofu-agent/conf.d

mkdir -p %{buildroot}/%{_sysconfdir}/sysconfig
cp %{SOURCE2} %{buildroot}/%{_sysconfdir}/sysconfig/cofu-agent

mkdir -p %{buildroot}/var/lib/cofu-agent

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
mkdir -p %{buildroot}/%{_unitdir}
cp %{SOURCE3} %{buildroot}/%{_unitdir}/
%endif

%pre

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
%post
%systemd_post cofu-agent.service
systemctl daemon-reload

%preun
%systemd_preun cofu-agent.service
systemctl daemon-reload
%endif

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%attr(755, root, root) %{_bindir}/%{name}
%dir %attr(755, root, root) /var/lib/cofu-agent
%config(noreplace) %{_sysconfdir}/cofu-agent/cofu-agent.toml
%config(noreplace) %{_sysconfdir}/sysconfig/cofu-agent
%dir %attr(755, root, root) %{_sysconfdir}/cofu-agent/conf.d

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
%{_unitdir}/cofu-agent.service
%endif

%doc
