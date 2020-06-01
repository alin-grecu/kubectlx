# kubectlx

Basically [kubectx](https://github.com/ahmetb/kubectx) for [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/)

```bash
kubectlx {version}
```

All versions are saved in `/usr/local/bin/kubectl-{version}`

## Install
* Download the latest [release](https://github.com/alin-grecu/kubectlx/releases/latest)

* Move binary to `/usr/local/bin`:
```bash
mv kubectlx /usr/local/bin
```

* Set execute permissions:
```bash
chmod +x /usr/local/bin/kubectlx
```

## Usage

```bash
kubectlx --help
```

## Example output:

### No kubectl installed:
```
$ kubectlx 1.12.0
2020/05/30 18:16:30 You do not have this version. Do you want to download it? [y/n]: y
2020/05/30 18:16:37 You are using kubectl 1.12.0
$ kubectl version --client=true --short
Client Version: v1.12.0
```

### Same version as the current one:
```
$ kubectl version --client=true --short
Client Version: v1.12.0
$ kubectlx 1.12.0
2020/05/30 18:17:08 Saved current version under /usr/local/bin/kubectl-1.12.0
2020/05/30 18:17:08 You are using kubectl 1.12.0
```

### Missing kubectl version:
```
$ kubectlx 1.18.0
2020/05/30 18:20:40 Saved current version under /usr/local/bin/kubectl-1.12.0
2020/05/30 18:20:40 You do not have this version. Do you want to download it? [y/n]: y
2020/05/30 18:20:48 You are using kubectl 1.18.0
$ kubectl version --client=true --short
Client Version: v1.18.0
```

### Existing kubectl version:
```
$ kubectl version --client=true --short
Client Version: v1.18.0
$ kubectlx 1.12.0
2020/05/30 18:25:18 Saved current version under /usr/local/bin/kubectl-1.18.0
2020/05/30 18:25:18 You are using kubectl 1.12.0
```
