using System.Net;
using System.Net.WebSockets;
using System.Reflection;
using System.Text;
using fun.dataType.Attribute.Bean;
using fun.dataType.Attribute.Service;
using fun.dataType.Service;
namespace fun;

public class Fun
{
    private  Connect _connect ;
    private  Dictionary<string, ServiceMethod> _serviceMethodMap = new();

    public async Task Run(ushort port)
    {
        Scan();
        _connect = new Connect(port);
        Console.WriteLine("WebSocket Server Started...");
        ThreadPool.SetMaxThreads(Environment.ProcessorCount,Environment.ProcessorCount);
        while (true)
        {
            var ws = await _connect.GetConnect();
            ThreadPool.QueueUserWorkItem(o => GetMessage(ws));
        }
    }
    
    private async Task GetMessage(Ws ws)
    {
        while (true)
        {
            try
            {
                var message = await ws.GetMessage();
                ws.Send(Result.Success("1111"));
                Console.WriteLine(message); 
            }
            catch 
            {   
                break;  
            }
        }
    }

    private void Scan()
    {
        var assembly = Assembly.GetEntryAssembly();
        var classes = assembly?.GetTypes();
        if (classes == null) return;
        foreach (var item in classes)
        {
            if (item.GetCustomAttribute<Service>() != null)
            {
                AddService(item);
            }
            else if (item.GetCustomAttribute<Bean>() != null)
            {
                // 处理Bean属性的类
            }
        }
    }

    private void AddService(Type service)
    {
        if (!service.IsSubclassOf(typeof(Ctx))) throw new ApplicationException("Service is not a subclass of Ctx");
        var serviceMethod = new ServiceMethod
        {
            ServiceType = service
        };
        _serviceMethodMap.Add(service.Name, serviceMethod);
        foreach (var method in service.GetMethods()) AddMethod(serviceMethod, method);
    }

    private void AddMethod(ServiceMethod serviceMethod, MethodInfo method)
    {
        string[] systemMethods = { "GetType", "ToString", "Equals", "GetHashCode" };
        if (!systemMethods.Contains(method.Name))
        {
            if (method.GetParameters().Length > 1)
                throw new ApplicationException("Method parameter count is more than 1");
            serviceMethod.MethodMap.Add(method.Name, method);
        }
    }
}

