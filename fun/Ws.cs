using System.Net.WebSockets;
using System.Text;
using fun.dataType.Service;

namespace fun;
using System.Net;
using System.Text.Json;



public class Ws
{
    public WebSocket WebSocket;
    public HttpListenerContext Ctx;
    private CancellationTokenSource _cts;
    private List<ArraySegment<byte>> _messageParts;
    private Byte[] _buffer = new Byte[1024 * 4];
    public Ws()
    {
        _cts = new CancellationTokenSource();
        _messageParts = new List<ArraySegment<byte>>();
        ExecuteWithDelayAsync(WebSocket, 2000, _cts.Token);
    }
    public async Task<String> GetMessage1()
    {
        var result = await WebSocket.ReceiveAsync(new ArraySegment<byte>(_buffer), CancellationToken.None);
        if (result.MessageType == WebSocketMessageType.Close)
        {
            Console.WriteLine("Client closed the WebSocket connection.");
            if (WebSocket.State == WebSocketState.Open) await Close();
            Ctx.Response.Close();
            throw new WebSocketException("WebSocket connection closed.");
        }

        if (result.MessageType == WebSocketMessageType.Text)
        {
            // 处理接收到的文本消息
            _messageParts.Add(new ArraySegment<byte>(_buffer, 0, result.Count));

            if (result.EndOfMessage)
            {
                // 将所有消息部分合并成一个完整的消息
                var completeMessageBuffer = CombineSegments(_messageParts);
                _messageParts.Clear();
                var message = Encoding.UTF8.GetString(completeMessageBuffer);
                if (!message.Equals("0"))
                {
                    return message;
                }
                _cts.Cancel();
                _cts.Dispose();
                _cts = new CancellationTokenSource();
                await Send(1);
                ExecuteWithDelayAsync(WebSocket, 7000, _cts.Token);
            }
        }
        return null;
    }

    public async Task Close()
    {
        await WebSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing as requested by server",
            CancellationToken.None);
    }
    
    public async Task Send(Object message)
    {
        await WebSocket.SendAsync(new ArraySegment<byte>(Encoding.UTF8.GetBytes(JsonSerializer.Serialize(message))), WebSocketMessageType.Text, true, CancellationToken.None);
    }
    
    public async Task<String> GetMessage()
    {
        String message = null;
        while (message == null)
        {
            message = await GetMessage1();
        }
        return message;
    }
    
    private byte[] CombineSegments(List<ArraySegment<byte>> segments)
    {
        var totalLength = segments.Sum(segment => segment.Count); // 这里只能在接收完整消息时计算一次总长度
        var completeBuffer = new byte[totalLength];
        var offset = 0;
        foreach (var segment in segments)
        {
            Buffer.BlockCopy(segment.Array, segment.Offset, completeBuffer, offset, segment.Count);
            offset += segment.Count;
        }

        return completeBuffer;
    }
    internal  async Task ExecuteWithDelayAsync(WebSocket websocket, ushort delayTime, CancellationToken cancellationToken)
    {
        await Task.Delay(delayTime, cancellationToken);
        await Close();
    }
}

