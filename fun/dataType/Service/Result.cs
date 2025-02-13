namespace fun.dataType.Service;
public class Result
{
    public string Id { get; set; }
    public ushort? Code{ get; set; }
    public  object Data{ get; set; }
    public ResultStatusEnum Status{ get; set; }
    public static Result Success(object data)
    {
        return new Result
        {
            Data = data,
            Status = ResultStatusEnum.Success
        };
    }

    public static Result Error(ushort code, object data)
    {
        return new Result
        {
            Code = code,
            Data = data,
            Status = ResultStatusEnum.Error
        };
    }

    internal static Result CallError(object data)
    {
        return new Result
        {
            Data = data,
            Status = ResultStatusEnum.CellError
        };
    }
}