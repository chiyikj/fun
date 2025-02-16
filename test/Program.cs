// See https://aka.ms/new-console-template for more information

using fun.dataType.Service;
using fun.dataType.Attribute.Service;
using fun;
using System.Reflection;
Fun fun = new Fun();
String a = "Hello World";
Console.WriteLine(a.GetType());
await fun.Run(3000);

[Service]
class MyClass1:Ctx
{
    
}

[Service]
class MyClass:Ctx
{
    public  int Add(int a)
    {
        return a;
    }
}


class anotherClass<T>
{
    public T A;
}