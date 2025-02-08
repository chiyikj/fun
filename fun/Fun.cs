namespace fun;
using System.Reflection;
using System.Net;
using System.Net.WebSockets;
using System.Text;
using fun.dataType.Attribute.Service;
using fun.dataType.Attribute.Bean;
using fun.dataType.Service;
public class Fun {
    private Dictionary<string, ServiceMethod> _serviceMethodMap = new Dictionary<string, ServiceMethod>();
    private  HttpListener _listener =  new HttpListener();
    private  CancellationTokenSource _cancellationTokenSource = new CancellationTokenSource();
    public async Task Run(ushort port) {
        Scan();
        _listener.Prefixes.Add($"http://*:{port}/");
        _listener.Start();
        Console.WriteLine("WebSocket Server Started...");

        while (!_cancellationTokenSource.Token.IsCancellationRequested) {
            var ctx = await _listener.GetContextAsync();
            if (ctx.Request.IsWebSocketRequest)
            {
                Console.WriteLine("WebSocket connection accepted");
                await AcceptWebSocketAsync(ctx);
            }
            else
            {
                Console.WriteLine("WebSocket connection rejected");
                ctx.Response.StatusCode = 400;
                ctx.Response.Close();
            }
        }
        _listener.Stop();
        Console.WriteLine("WebSocket Server stopped...");
    }

    private void Scan() {
        var assembly = Assembly.GetEntryAssembly();
        var classes = assembly?.GetTypes();
        if (classes == null) return;
        foreach (var item in classes)
        {
            if (item.GetCustomAttribute<Service>() != null) {
                AddService(item);
            } else if (item.GetCustomAttribute<Bean>() != null) {

            }
        }
    }

    private void AddService(Type service) {
        if (!service.IsSubclassOf(typeof(Ctx))) {
            throw new ApplicationException("Service is not a subclass of Ctx");
        }
        ServiceMethod serviceMethod = new ServiceMethod
        {
            ServiceType = service,
        };
        _serviceMethodMap.Add(service.Name, serviceMethod);
        foreach (var method in service.GetMethods()) {
            AddMethod(serviceMethod,method);
        }
    }

    private void AddMethod(ServiceMethod serviceMethod,MethodInfo method)
    {
        string[] systemMethods = { "GetType", "ToString", "Equals", "GetHashCode" };
        if (!systemMethods.Contains(method.Name)) {
            if (method.GetParameters().Length > 1) {
                throw new ApplicationException("Method parameter count is more than 1");
            }
            serviceMethod.MethodMap.Add(method.Name, method);
        }
    }
    private async Task AcceptWebSocketAsync(HttpListenerContext context) {
        var webSocketContext = await context.AcceptWebSocketAsync(null);
        var webSocket = webSocketContext.WebSocket;
        var buffer = new byte[1024 * 4]; // 固定大小的缓冲区用于接收每个数据段
        var messageParts = new List<ArraySegment<byte>>();
        while (webSocket.State == WebSocketState.Open) {
            var result = await webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), _cancellationTokenSource.Token);

            if (result.MessageType == WebSocketMessageType.Close) {
                await webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing as requested by client", _cancellationTokenSource.Token);
                break;
            }
            if (result.MessageType == WebSocketMessageType.Text) {
                // 处理接收到的文本消息
                messageParts.Add(new ArraySegment<byte>(buffer, 0, result.Count));

                if (result.EndOfMessage) {
                    // 将所有消息部分合并成一个完整的消息
                    var completeMessageBuffer = CombineSegments(messageParts);
                    var message = Encoding.UTF8.GetString(completeMessageBuffer);
                    Console.WriteLine("Received message from client: " + message);
                    Console.WriteLine("Message size: " + completeMessageBuffer.Length + " bytes");

                    // 将消息原样发送回客户端
                    await webSocket.SendAsync(new ArraySegment<byte>(completeMessageBuffer), result.MessageType, result.EndOfMessage, _cancellationTokenSource.Token);

                    // 清空消息部分列表，准备接收下一个消息
                    messageParts.Clear();
                }
            }
        }
    }

    private byte[] CombineSegments(List<ArraySegment<byte>> segments) {
        var totalLength = segments.Sum(segment => segment.Count);
        var completeBuffer = new byte[totalLength];
        int offset = 0;

        foreach (var segment in segments) {
            Buffer.BlockCopy(segment.Array, segment.Offset, completeBuffer, offset, segment.Count);
            offset += segment.Count;
        }

        return completeBuffer;
    }
}








