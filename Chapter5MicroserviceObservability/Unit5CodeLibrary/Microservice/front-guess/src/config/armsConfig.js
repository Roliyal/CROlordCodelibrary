// armsConfig.js

export const armsConfig = {
    // 必填项
    pid: "djqtzchc9t@a247bc2e041fd12",            // 应用ID，替换为你的实际应用ID
    endpoint: "https://djqtzchc9t-default-sea.rum.aliyuncs.com",      // 上报数据的端点，替换为你的实际上报地址

    // 可选项
    env: "prod",                    // 应用环境：prod、gray、pre、daily、local（默认为prod）
    version: "1.0.0",               // 应用版本号
    user: {                         // 用户信息配置
        id: "user_id",                // 用户ID，SDK默认生成，或者可以自定义
        name: "user_name",            // 用户名称
        tags: "user_tags",            // 用户标签
    },
    tracing: {
        enable: true, // 开启链路追踪，默认关闭
        sample: 100, // 采样率，默认100%
        tracestate: true, // 开启tracestate透传，默认开启
        baggage: true, // 开启baggage透传，默认关闭
        allowedUrls:[
            {match: 'https://micro.roliyal.com', propagatorTypes:['tracecontext', 'b3']}, // 字符匹配 https://api.aliyun.com开头，使用w3c标准
            {match: /micro\.roliyal\.com/i, propagatorTypes:['b3multi']}, // 正则匹配包含roliyal，使用b3multi多头标准
            {match: (url)=>url.includes('.api'), propagatorTypes:['jaeger']}, // 函数判断包含.api， 使用jaeger标准
        ]
    },
    // 上报配置
    reportConfig: {
        flushTime: 3000,              // 上报时间间隔，单位：毫秒
        maxEventCount: 50             // 每次上报的最大事件数
    },

    // Session 配置
    sessionConfig: {
        sampleRate: 0.5,              // Session采样率：0~1之间的数值，表示采样的比例
        maxDuration: 86400000,        // Session最大持续时间，单位：毫秒（默认24小时）
        overtime: 3600000,            // Session超时，单位：毫秒（默认1小时）
        storage: 'localStorage'       // Session存储位置，可以是 'cookie' 或 'localStorage'
    },

    // 采集器配置
    collectors: {
        action: true,                 // 是否监听用户行为（例如点击事件）
        api: true,                    // 是否监听API请求
        jsError: true,                // 是否监听JS错误
        consoleError: true,           // 是否监听Console错误
        perf: true,                   // 是否监听页面性能
        staticResource: true,         // 是否监听静态资源请求
        webvitals: true               // 是否监听Web Vitals数据
    },

    // 白屏监控配置
    whiteScreen: {
        detectionRules: [{
            target: '#root',             // 目标元素的选择器
            test_when: ['LOAD', 'ERROR'],// 触发事件：页面加载（LOAD）、发生错误（ERROR）
            delay: 5000,                 // 延时5秒开始检测
            tester: 'SCREENSHOT',        // 使用截图法进行白屏检测
            configOptions: {
                colorRange: ['rgb(255, 255, 255)'], // 用于像素比对的颜色集合
                threshold: 0.9,            // 白屏率阈值，大于该值认为是白屏
                pixels: 10,                // 检测区域的像素大小
                horizontalOffset: 210,     // 水平偏移量
                verticalOffset: 50        // 垂直偏移量
            }
        }]
    },

    // 动态配置
    remoteConfig: {
        region: "ap-southeast-1"         // 配置所在的Region，例如：ap-southeast-1、cn-hangzhou等
    },

    // 自定义属性配置
    properties: {
        prop_string: 'example',      // 字符串类型的自定义属性
        prop_number: 123,            // 数字类型的自定义属性
        prop_boolean: true,          // 布尔类型的自定义属性
        more_than_50_key_limit_012345678901234567890123456789: 'long-key-value', // 键名不能超过50个字符
        more_than_2000_value_limit: new Array(2003).join('1') // 值不能超过2000字符
    },

    // 事件过滤配置
    filters: {
        exception: [
            'Test error',                // 过滤以'Test error'开头的异常信息
            /^Script error\.?$/,         // 使用正则表达式匹配异常信息
            (msg) => msg.includes('example-error') // 自定义过滤函数
        ],
        resource: [
            'https://example.com/',      // 过滤以'https://example.com/'开头的资源请求
            /localhost/i,                // 正则匹配localhost
            (url) => url.includes('example-resource') // 自定义函数进行过滤
        ]
    },

    // 自定义API解析配置
    evaluateApi: async (options, response, error) => {
        let respText = '';
        if (response && response.text) {
            respText = await response.text();
        }
        return {
            name: 'my-custom-api',       // 自定义API名称
            success: error ? 0 : 1,      // 请求成功状态：0表示失败，1表示成功
            snapshots: JSON.stringify({
                params: 'page=1&size=10', // 请求参数
                response: respText.substring(0, 2000), // 响应内容（截取前2000字符）
                reqHeaders: '',           // 请求头
                resHeaders: ''            // 响应头
            }),
            properties: {
                custom_prop: 'custom_value'  // 自定义属性
            }
        };
    },

    // 地理信息配置
    geo: {
        country: 'your  country info', // 自定义国家信息
        city: 'your custom city info'        // 自定义城市信息
    }
};
