package framework

type commandOpt func(*[]string)

func appendOpt(new ...string) commandOpt {
	return func(args *[]string) {
		*args = append(*args, new...)
	}
}

func withKubeconfig(kubeconfigFile string) commandOpt {
	return appendOpt("--kubeconfig", kubeconfigFile)
}
