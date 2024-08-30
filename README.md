# techxsort
# why we need to use this tool
```
#run techx with directorys also so you won't miss anything, bacause "chrome Wappalyzer" extension and techx uses .js .css file so each directory may contains different .js, .css files, if you won't run like this you will miss some of results, example shown
# do not run nuclei info template for detecting technologies, nuclei will run all info templates website even not using these tech you running with these templates
# you don't need to run on all urls just run domain and 5th level directory paths, like this
https://devportal.radcns.com
https://devportal.radcns.com/jenkins
https://devportal.radcns.com/jenkins/fashion
https://devportal.radcns.com/jenkins/fashion/script

echo "https://devportal.radcns.com" | techx
URL: https://devportal.radcns.com
Count: 5
Technologies: [Nginx:1.26.0, PHP:5.5.12, Ubuntu, jQuery, jQuery CDN]

root@DESKTOP-7JCSD20:~# echo "https://devportal.radcns.com/jenkins" | techx
URL: https://devportal.radcns.com/jenkins
Count: 1
Technologies: [Nginx:1.26.0]

root@DESKTOP-7JCSD20:~# echo "https://devportal.radcns.com/jenkins/fashion" | techx
URL: https://devportal.radcns.com/jenkins/fashion
Count: 5
Technologies: [Java, Jenkins:2.375.2, Nginx:1.26.0, Prototype, YUI]


# nuclei, missed with the main domain and other directory paths, and got 1 jenkins rce because scanned on all directory paths
echo "https://devportal.radcns.com/jenkins/fashion" | nuclei -duc -t cent-nuclei-templates/ -tags jenkins
[unauthenticated-jenkins-rce] [http] [critical] https://devportal.radcns.com/jenkins/fashion/script/
```

techxsort extracts tech in the domain and 5th level directory paths and count all tech then print with the domain and removes all directory paths, that means the website using all of this techs

# Installation
```
go install github.com/rix4uni/techxsort@latest
```

##### via clone command
```
git clone https://github.com/rix4uni/techxsort.git && cd techxsort && go build techxsort.go && mv techxsort ~/go/bin/techxsort && cd .. && rm -rf techxsort
```

##### Usage
```
Usage of techxsort:
  -file string
        Path to the techx JSON file
  -o string
        Path to save the output file
```

# Output Examples

Single URL:
```
echo "https://carson.math.uwm.edu" | techxsort -file techx-output.json
```

Multiple URLs:
```
cat httpx.txt | techxsort -file techx-output.json
```

```
cat httpx.txt
https://carson.math.uwm.edu
https://www.shodan.io
https://agilezws.us.dell.com
https://math.uwm.edu
```

techxsort only supports json input like this `techx-output.json`
```
cat techx-output.json
{
  "host": "https://agilezws.us.dell.com",
  "count": 1,
  "tech": [
    "df"
  ]
}
{
  "host": "https://carson.math.uwm.edu",
  "count": 2,
  "tech": [
    "Apache HTTP Server",
    "HSTS"
  ]
}
{
  "host": "https://carson.math.uwm.edu/jenkins/computer/(built-in)/script",
  "count": 6,
  "tech": [
    "Apache HTTP Server",
    "HSTS",
    "Java",
    "Jenkins:2.452.2",
    "Jetty:10.0.20",
    "YUI"
  ]
}
```

# JSON format
```
cat httpx.txt | techxsort -file techx-output.json
{
  "host": "https://carson.math.uwm.edu",
  "count": 6,
  "tech": [
    "Apache HTTP Server",
    "HSTS",
    "Java",
    "Jenkins:2.452.2",
    "Jetty:10.0.20",
    "YUI"
  ]
}
```

# Usage Example
```
subfinder -d hackerone.com -all -duc -silent | httpx -duc -silent -mc 200 -t 300 | unew httpx.txt
cat httpx.txt | katana -duc -silent -f udir -ct 60 | unew | techx -json -o techx-output.json
cat httpx.txt | techxsort -file techx-output.json -o techxsort.json
nucleitechx -file techxsort.json -nucleicmd "nuclei -duc -silent -tags {tech} -es unknown,info,low" -process -append nuclei-output.txt
```
