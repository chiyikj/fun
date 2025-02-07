namespace fun;

public class Bean : Attribute
{
    public BeanType beanType { get; private set; }

    // 默认构造函数，设置默认的 BeanType 为 Single
    public Bean()
    {
        this.beanType = BeanType.Single;
    }

    // 带参数的构造函数，允许指定 BeanType
    public Bean(BeanType beanType)
    {
        this.beanType = beanType;
    }
}