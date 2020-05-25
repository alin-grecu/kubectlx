# kubectlx

Basically kubectx for kubectl

```bash
kubectlx {version}
```

All versions are saved in `/usr/local/bin/kubectl-{version}`

## Example output:

### Missing kubectl version:
```bash
kubectlx 1.11.6
2020/05/25 19:39:49 stat /usr/local/bin/kubectl-1.11.6: no such file or directory
2020/05/25 19:39:49 Do you want to download this version? [y/n]:
y
2020/05/25 19:39:52 Downloading kubectl version 1.11.6
```

### Same version as the current one:
```bash
kubectlx 1.11.6
2020/05/25 19:40:45 You are already using version 1.11.6
```
