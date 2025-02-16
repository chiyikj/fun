namespace fun;
using System.Reflection;
public class Check
{
    internal static void  IsBasicType(Type type)
    {
        var typeList = new List<Type>
        {
            typeof(bool), typeof(byte), typeof(int), typeof(long), typeof(short), typeof(sbyte), typeof(uint),
            typeof(ulong), typeof(ushort), typeof(string)
        };
        if (!typeList.Contains(type) && !type.IsEnum)
        {
            throw new ArgumentException("Type is not a basic type");  
        }
    }

    internal static void IsType(Type type)
    {
        if (type.IsClass)
        {
            FieldInfo[] fields = type.GetFields(BindingFlags.Public | BindingFlags.Instance);
            foreach (var field in fields)
            {
                IsType(field.FieldType);
            }
        }
        if (type == typeof(List<>))
        {
            IsType(type.GetGenericArguments()[0]);
        }
        IsBasicType(type);
    }
    
    internal static void IsParameter(ParameterInfo parameter)
    {
        if
    }
    
    internal static void IsReturn(Type returnType)
    {
        IsType(returnType);
    }
}