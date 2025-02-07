// See https://aka.ms/new-console-template for more information

using fun;
using fun.dataType.Service;

Fun fun = new Fun();
await fun.Run(3000);

[Service]
class MyClass1:Ctx
{
    
}

[Service]
class MyClass:Ctx
{
    public async Task<int> add(int a)
    {
        return a;
    }
}
