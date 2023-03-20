export default {
  async login(username, password) {
    const response = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ username, password }),
    });


    if (!response.ok) {
      throw new Error("登录失败");
    }
  },
};
