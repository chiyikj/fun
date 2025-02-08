namespace fun.dataType.Service;
using System.Reflection;
internal class ServiceMethod
{
    public Dictionary<string, MethodInfo> MethodMap = new Dictionary<string, MethodInfo>();
    public required Type ServiceType;

}