namespace fun.dataType.Attribute.Service;
using System;

    
public class Proxy : Attribute
{
    private Type _type;
    public Proxy(Type type)
    {
        _type = type;
    }
}
