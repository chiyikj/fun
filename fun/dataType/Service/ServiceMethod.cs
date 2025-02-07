namespace fun.dataType.Service;
using System.Reflection;
class ServiceMethod
{
    public Dictionary<string, MethodInfo> methodMap = new Dictionary<string, MethodInfo>();
    public Type serviceType;

}