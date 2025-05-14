export const createArmsConfig = (userId) => {
    return {
        // 必填项
        pid: "djqtzchc9t@a247bc2e041fd12",  // 应用ID，替换为你的实际应用ID
        endpoint: "https://djqtzchc9t-default-sea.rum.aliyuncs.com",  // 上报数据的端点，替换为你的实际上报地址

        // 可选项
        env: "prod",  // 应用环境：prod、gray、pre、daily、local（默认为prod）
        version: "1.0.0",  // 应用版本号
        user: {  // 用户信息配置
            id: "user_id",  // 用户ID，SDK默认生成，不支持更改
            //name: "user_name",            // 用户名称
            name: userId, // 用户名称可以设置为业务中的 userId
            tags: "User_Level_Demo\n",  // 用户属性演示
        },
        tracing: {
            enable: true,  // 开启链路追踪，默认关闭
            sample: 100,  // 采样率，默认100%
            tracestate: true,  // 开启tracestate透传，默认开启
            baggage: true,  // 开启baggage透传，默认关闭
            allowedUrls:[ // 配置需要透传的协议 URL ，根据实际需求选择
                {match: 'https://micro.roliyal.com', propagatorTypes:['tracecontext', 'b3']}, // 字符匹配 https://api.aliyun.com开头，使用w3c标准
                {match: /micro\.roliyal\.com/i, propagatorTypes:['b3multi']}, // 正则匹配包含roliyal，使用b3multi多头标准
                {match: (url)=>url.includes('.api'), propagatorTypes:['jaeger']}, // 函数判断包含.api， 使用jaeger标准
            ]
        },
        // 上报配置
        reportConfig: {
            flushTime: 3000,
            maxEventCount: 50,
        },

        // Session 配置
        sessionConfig: {
            sampleRate: 0.5,
            maxDuration: 86400000,
            overtime: 3600000,
            storage: 'localStorage',
        },

        // 采集器配置
        collectors: {
            action: true,
            api: true,
            jsError: true,
            consoleError: true,
            perf: true,
            staticResource: true,
            webvitals: true,
        },

        // 白屏监控配置
        whiteScreen: {
            detectionRules: [{
                target: '#root',  // 目标元素的选择器
                test_when: ['LOAD', 'ERROR'],  // 触发事件：页面加载（LOAD）、发生错误（ERROR）
                delay: 5000,  // 延时5秒开始检测
                tester: 'SCREENSHOT',  // 使用截图法进行白屏检测
                configOptions: {
                    colorRange: ['rgb(255, 255, 255)'],  // 用于像素比对的颜色集合
                    threshold: 0.9,  // 白屏率阈值，大于该值认为是白屏
                    pixels: 10,  // 检测区域的像素大小
                    horizontalOffset: 210,  // 水平偏移量
                    verticalOffset: 50  // 垂直偏移量
                }
            }]
        },

        // 动态配置
        // remoteConfig: {
        //     region: "ap-southeast-1"  // 配置所在的Region，例如：ap-southeast-1、cn-hangzhou等
        // },

        // 自定义属性配置
        properties: {
            is_logged_in: true,
            ser_level: 'premium',
            app_version: "1.0.0"
        },

        // 事件过滤配置
        filters: {
            exception: [
                'Test error',  // 过滤以'Test error'开头的异常信息
                /^Script error\.?$/,  // 使用正则表达式匹配异常信息
                (msg) => msg.includes('example-error')  // 自定义过滤函数
            ],
            resource: [
                'https://example.com/',  // 过滤以'https://example.com/'开头的资源请求
                /localhost/i,  // 正则匹配localhost
                (url) => url.includes('example-resource')  // 自定义函数进行过滤
            ]
        },

        // 自定义API解析配置
        evaluateApi: async (options, response, error) => {
            let respText = '';
            if (response && response.text) {
                respText = await response.text();
            }

            const apiName = options.url.split('/').pop();  // 获取 URL 中最后一个部分，作为 API 名称

            return {
                name: apiName,  // 使用动态生成的 API 名称
                success: error ? 0 : 1,  // 请求成功状态，0表示失败，1表示成功
                snapshots: JSON.stringify({
                    params: options.params || '',  // 请求参数
                    response: respText.substring(0, 2000),  // 响应内容（截取前2000字符）
                    reqHeaders: JSON.stringify(options.headers || {}),  // 请求头
                    resHeaders: JSON.stringify(response.headers || {})  // 响应头
                }),
                properties: {
                    user_id: userId,  // 当前用户ID
                    //api_type: apiType  // API 类型（用于区分用户相关和通用API）
                }
            };
        },
 // 在数据上报之前，获取 traceId 并添加到 properties
        beforeReport: (reportData) => {
            console.log("Before report data:", JSON.stringify(reportData, null, 2));

            let traceId = 'No traceId available';

            // 查找 events 中的 trace_data 并提取 X-B3-TraceId
            if (reportData && reportData.events && Array.isArray(reportData.events)) {
                for (let i = 0; i < reportData.events.length; i++) {
                    const event = reportData.events[i];

                    // 直接从 trace_data 中获取 X-B3-TraceId
                    if (event.trace_data && event.trace_data.headers) {
                        traceId = event.trace_data.headers['X-B3-TraceId'] || traceId;
                        break; // 找到 trace_id 后退出循环
                    }
                }
            }

            // 将 traceId 添加到 properties
            reportData.properties = {
                ...reportData.properties,
                'X-B3-TraceId': traceId,
            };

            console.log("Trace ID from events:", traceId);

            return reportData;
        }
        // 地理信息配置
        //geo: {
        //    country: 'your country info',  // 自定义国家信息
        //    city: 'your custom city info'  // 自定义城市信息
        //}
    };
};
