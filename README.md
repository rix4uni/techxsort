## techxsort
Deduplicates technologies across hosts sharing the same domain. Preserves individual host paths but ensures each technology only appears once per domain — so nuclei doesn't scan the same tech twice.

# How it works
1. Groups hosts by domain (hostname only, ignores path)
2. Within each domain group, processes hosts by highest tech count first (first-seen tiebreak for equal counts)
3. Each tech is "claimed" by the first host processed — removed from later hosts
4. Hosts left with 0 remaining unique techs are removed entirely

## Installation
### Install via Go
```
go install github.com/rix4uni/techxsort@latest
```

### Download Prebuilt Binaries
```
wget https://github.com/rix4uni/techxsort/releases/download/v0.0.2/techxsort-linux-amd64-0.0.2.tgz
tar -xvzf techxsort-linux-amd64-0.0.2.tgz
rm -rf techxsort-linux-amd64-0.0.2.tgz
mv techxsort ~/go/bin/techxsort
```

Or download the [latest release](https://github.com/rix4uni/techxsort/releases) for your platform.

### Compile from Source
```
git clone --depth 1 https://github.com/rix4uni/techxsort.git
cd techxsort; go install
```

##### Usage
```console
Usage of techxsort:
  -o string
        Path to save the output file
  -version
        Print the version of the tool and exit.
```

# Output Examples
```console
cat techx-output.json | techxsort
cat techx-output.json | techxsort -o filtered.json
```

# Example
```console
# input
cat techx-output.json
{
  "host": "https://devportal.radcns.com",
  "count": 5,
  "tech": [
    "Nginx:1.26.0",
    "PHP:5.5.12",
    "Ubuntu",
    "jQuery",
    "jQuery CDN"
  ]
}
{
  "host": "https://devportal.radcns.com/jenkins",
  "count": 1,
  "tech": [
    "Nginx:1.26.0"
  ]
}
{
  "host": "https://devportal.radcns.com/jenkins/fashion",
  "count": 5,
  "tech": [
    "Java",
    "Jenkins:2.375.2",
    "Nginx:1.26.0",
    "Prototype",
    "YUI"
  ]
}

# output (Nginx:1.26.0 claimed by root, /jenkins removed entirely, /jenkins/fashion keeps unique techs)
cat techx-output.json | techxsort
{
  "host": "https://devportal.radcns.com",
  "count": 5,
  "tech": [
    "Nginx:1.26.0",
    "PHP:5.5.12",
    "Ubuntu",
    "jQuery",
    "jQuery CDN"
  ]
}
{
  "host": "https://devportal.radcns.com/jenkins/fashion",
  "count": 4,
  "tech": [
    "Java",
    "Jenkins:2.375.2",
    "Prototype",
    "YUI"
  ]
}
```

# Usage Example
```console
subfinder -d hackerone.com -all -duc -silent | httpx -duc -silent -mc 200 -t 300 | unew httpx.txt
cat httpx.txt | katana -duc -silent -fs fqdn -c 100 -p 100 -f udir -ct 60 -o katana.txt
cat katana.txt | unew | techx -json -o techx-output.json
cat techx-output.json | techxsort -o techxsort-output.json
nucleitechx -file techxsort-output.json -nucleicmd "nuclei -duc -silent -tags {tech} -es unknown,info,low" -process -append nuclei-output.txt
```
