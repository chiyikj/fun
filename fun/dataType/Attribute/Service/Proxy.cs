namespace fun.dataType.Service;

public class Proxy : Attribute
{
    public Type type;

    public Proxy(Type type)
    {
        this.type = type;
    }
}