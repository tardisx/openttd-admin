#!/usr/bin/env perl

use strict;
use warnings;

open my $fh, "<", "cmd/openttd_multitool/openttd_multitool.go" || die $!;

my $version;
while (<$fh>) {
  $version = $1 if /^const\s+currentVersion.*?"([\d\.]+)"/;
}
close $fh;

die "no version?" unless defined $version;

# so lazy
system "rm", "-rf", "release", "dist";
system "mkdir", "release";
system "mkdir", "dist";

my %build = (
  win   => { env => { GOOS => 'windows', GOARCH => 'amd64' }, filename => 'openttd_multitool.exe' },
  linux => { env => { GOOS => 'linux',   GOARCH => 'amd64' }, filename => 'openttd_multitool' },
  mac   => { env => { GOOS => 'darwin',  GOARCH => 'amd64' }, filename => 'openttd_multitool' },
);

foreach my $type (keys %build) {
  mkdir "release/$type";
}

foreach my $type (keys %build) {
  local $ENV{GOOS}   = $build{$type}->{env}->{GOOS};
  local $ENV{GOARCH} = $build{$type}->{env}->{GOARCH};
  system "go", "build", "-o", "release/$type/" . $build{$type}->{filename}, "cmd/openttd_multitool/openttd_multitool.go";
  system "zip", "-j", "dist/openttd_multitool-$type-$version.zip", ( glob "release/$type/*" );
}
