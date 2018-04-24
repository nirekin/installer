# Building the Installer image

The image can be built using the script `build.sh`

This script accepts two parameters

- First parameter: the ***http_proxy*** entry
- Second parameter: the ***https_proxy*** entry 

> If the ***https_proxy*** entry is the same than the ***http_proxy*** then the second parameter can be omitted.

Example:

```bash
./build.sh http://user:password@your.proxy.com:80
```


