using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.WebSockets;
using System.Reflection;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace fun {
    using fun.dataType.Attribute.Service;
    using fun.dataType.Attribute.Bean;
    using fun.dataType.Service;

    public class Fun {
        private Dictionary<string, ServiceMethod> _serviceMethodMap = new Dictionary<string, ServiceMethod>();
        private HttpListener _listener = new HttpListener();

        public async Task Run(ushort port) {
            Scan();
            _listener.Prefixes.Add($"http://*:{port}/");
            _listener.Start();
            Console.WriteLine("WebSocket Server Started...");
            while (true) {
                var ctx = await _listener.GetContextAsync();
                if (ctx.Request.IsWebSocketRequest) {
                    await AcceptWebSocketAsync(ctx);
                } else {
                    ctx.Response.StatusCode = 400;
                    ctx.Response.Close();
                }
            }
        }

        private void Scan() {
            var assembly = Assembly.GetEntryAssembly();
            var classes = assembly?.GetTypes();
            if (classes == null) return;
            foreach (var item in classes) {
                if (item.GetCustomAttribute<Service>() != null) {
                    AddService(item);
                } else if (item.GetCustomAttribute<Bean>() != null) {
                    // 处理Bean属性的类
                }
            }
        }

        private void AddService(Type service) {
            if (!service.IsSubclassOf(typeof(Ctx))) {
                throw new ApplicationException("Service is not a subclass of Ctx");
            }
            ServiceMethod serviceMethod = new ServiceMethod {
                ServiceType = service,
            };
            _serviceMethodMap.Add(service.Name, serviceMethod);
            foreach (var method in service.GetMethods()) {
                AddMethod(serviceMethod, method);
            }
        }

        private void AddMethod(ServiceMethod serviceMethod, MethodInfo method) {
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
            var cts = new CancellationTokenSource();

            // 启动延迟任务
            DelayExample example = new DelayExample();
            example.ExecuteWithDelayAsync(webSocket, cts.Token);

            var buffer = new byte[1024 * 4]; // 固定大小的缓冲区用于接收每个数据段
            var messageParts = new List<ArraySegment<byte>>();

            while (webSocket.State == WebSocketState.Open) {
                var result = await webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), CancellationToken.None);
                if (result.MessageType == WebSocketMessageType.Close) {
                    Console.WriteLine("Client closed the WebSocket connection.");
                    if (webSocket.State == WebSocketState.Open || webSocket.State == WebSocketState.CloseReceived) {
                        await webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing as requested by server", CancellationToken.None);
                    }
                    context.Response.Close();
                    break;
                }

                if (result.MessageType == WebSocketMessageType.Text) {
                    // 处理接收到的文本消息
                    messageParts.Add(new ArraySegment<byte>(buffer, 0, result.Count));

                    if (result.EndOfMessage) {
                        // 将所有消息部分合并成一个完整的消息
                        var completeMessageBuffer = CombineSegments(messageParts);
                        var message = Encoding.UTF8.GetString(completeMessageBuffer);
                        if (message.Equals("0")) {
                            cts.Cancel();
                            cts = new CancellationTokenSource();
                            await webSocket.SendAsync(new ArraySegment<byte>(Encoding.UTF8.GetBytes("1")), WebSocketMessageType.Text, true, CancellationToken.None);
                            DelayExample example1 = new DelayExample();
                            example1.ExecuteWithDelayAsync(webSocket, cts.Token);
                        } else {
                            await webSocket.SendAsync(new ArraySegment<byte>(completeMessageBuffer), result.MessageType, result.EndOfMessage, CancellationToken.None);
                        }
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

    public class DelayExample {
        public async Task ExecuteWithDelayAsync(WebSocket websocket, CancellationToken cancellationToken) {
            await Task.Delay(7000, cancellationToken);
            if (websocket.State == WebSocketState.Open || websocket.State == WebSocketState.CloseReceived) {
                await websocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing as requested by server", cancellationToken);
            }
        }
    }
}