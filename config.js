// 前端配置
window.appConfig = {
    // WebSocket连接地址
    signalingServer: window.location.origin.replace('http', 'ws') + '/ws',

    // API基础地址
    apiBaseUrl: '/api',

    // 媒体服务器地址
    mediaBaseUrl: '/media',

    // 默认房间ID
    defaultRoom: "math101",

    // UI配置
    ui: {
        theme: "light", // light/dark
        showParticipantList: true,
        enableScreenShare: true,
        enableChat: true
    },

    // 调试模式
    debug: false
};