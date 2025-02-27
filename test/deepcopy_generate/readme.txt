这里是一个根据types.go生成深拷贝方法的脚本，如果要生成深拷贝方法，首先需要将所有要用到的资源都放在此文件夹文件types.go中，
包括资源要调用的资源，不要分开若干个go文件，之后根据自动生成工具的要求在结构体前面进行注释补充

在package apis前面补充注释
// +k8s:deepcopy-gen=package

在需要访问的资源，如Node、NodeList等前面，需要补充注释
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object



部分情况下会出现如下问题：引用某些包比如time时，自动生成工具会将该结构体当作自行定义的
针对此问题，目前的解决方案是，要吗在生成之后自行修改
或者通过注释
// +k8s:deepcopy-gen=false
取消存在此问题的结构体，并在add_deepcopy.go文件中补充该结构体的深拷贝方法，之后运行单独补充生成。


通过命令运行脚本，生成的深拷贝方法存放在deepcopy.go文件中
./create_deepcopy.sh
将生成的文件替换掉deepcopy.go文件即可


