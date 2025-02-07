namespace fun.dataType.Service;

public delegate Task GuardDelegate(Result? result);
public interface IGuard
{
    // 接口成员
    public Task Use(Dictionary<string, object> state,string methodName,string serviceName,GuardDelegate guardDelegate);
}

