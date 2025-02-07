namespace fun.dataType.Service;
using System.Reflection;
internal class ServiceMethod
{
    public Dictionary<string, MethodInfo> methodMap = new Dictionary<string, MethodInfo>();
    public required Type ServiceType;

}