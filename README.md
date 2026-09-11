# Spring
Spring相关组件阅读笔记.

# 传送门

- [spring-core](https://github.com/seaswalker/Spring/blob/master/note/Spring.md)
- [spring-aop](https://github.com/seaswalker/Spring/blob/master/note/spring-aop.md)
- [spring-context](https://github.com/seaswalker/Spring/blob/master/note/spring-context.md)
- [spring-task](https://github.com/seaswalker/Spring/blob/master/note/spring-task.md)
- [spring-transaction](https://github.com/seaswalker/Spring/blob/master/note/spring-transaction.md)
- [spring-mvc](https://github.com/seaswalker/Spring/blob/master/note/spring-mvc.md)
- [guava-cache](https://github.com/seaswalker/Spring/blob/master/note/guava-cache.md)


# 运行 (Running the demonstrations)

Go 1.21 或更新版本, 无第三方依赖. 全部命令均从仓库根目录执行.

Go 1.21 or newer, no third-party dependencies. Run everything from the
repository root, so that `resources/` and `web/` resolve.

```sh
go build ./...        # build
go vet ./...          # vet
go test ./...         # run the test suite
```

| 命令 | 对应原类 | 说明 |
|---|---|---|
| `go run ./cmd/aop` | `aop.Bootstrap` | AOP 拦截顺序与 `expose-proxy` 自调用 |
| `go run ./cmd/base` | `base.Boostrap` | 事务 Bean 查找 (与原实现一样失败并以 1 退出) |
| `go run ./cmd/javaconfig` | `java_config.Bootrap` | 注解配置与 BeanDefinition 注册顺序 |
| `go run ./cmd/jdkproxy` | `test.proxy.JDKProxy` | 动态代理 |
| `go run ./cmd/javatest` | `test.JavaTest` | 反射 |
| `go run ./cmd/webapp` | `web.xml` + `spring-servlet.xml` | Spring MVC, http://localhost:8080/spring/ |
