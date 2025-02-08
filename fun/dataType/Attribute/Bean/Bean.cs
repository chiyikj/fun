namespace fun.dataType.Attribute.Bean;
using System;
using fun.dataType.Bean;

public class Bean : Attribute
{
    private BeanEnum _beanEnum;

    // 默认构造函数，设置默认的 BeanType 为 Single
    public Bean(): this(BeanEnum.Single)
    {
    }

    // 带参数的构造函数，允许指定 BeanType
    public Bean(BeanEnum beanType)
    {
        _beanEnum = beanType;
    }
}


