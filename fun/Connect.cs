namespace fun;
using System.Net;
public class Connect
{
    private readonly HttpListener _listener = new();
    public Connect(int port)
    {
        _listener.Prefixes.Add($"http://*:{port}/");
        _listener.Start();
    }
    public async Task<Ws> GetConnect()
    {
        Ws ws = null;
        while (ws == null)
        {
            ws = await AcceptWebSocket();
        }
        return ws;
    }
    
    private async Task<Ws?> AcceptWebSocket()
    {
        var ctx = await _listener.GetContextAsync();
        if (ctx.Request.IsWebSocketRequest)
        {
            var webSocketContext = await ctx.AcceptWebSocketAsync(null);
            return new Ws() { WebSocket = webSocketContext.WebSocket,Ctx = ctx };
        }
        ctx.Response.StatusCode = 400;
        ctx.Response.Close();
        return null;
    }
}