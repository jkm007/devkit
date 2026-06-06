package email

import "fmt"

// RenderVerifyCode 渲染验证码邮件 HTML
func RenderVerifyCode(code, purpose string, expireMinutes int, siteName string) string {
	if siteName == "" {
		siteName = "管理系统"
	}
	if expireMinutes == 0 {
		expireMinutes = 5
	}

	purposeText := purpose
	switch purpose {
	case "register":
		purposeText = "注册账号"
	case "reset_password":
		purposeText = "重置密码"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f5f5f5; padding: 40px 0;">
  <tr>
    <td align="center">
      <table width="480" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 12px rgba(0,0,0,0.08); overflow: hidden;">
        <!-- 头部 -->
        <tr>
          <td style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center;">
            <h1 style="color: #ffffff; margin: 0; font-size: 22px; font-weight: 600;">%s</h1>
          </td>
        </tr>
        <!-- 内容 -->
        <tr>
          <td style="padding: 40px 30px;">
            <p style="color: #333; font-size: 16px; margin: 0 0 20px;">您好，</p>
            <p style="color: #666; font-size: 14px; margin: 0 0 30px;">您正在进行<strong>%s</strong>操作，验证码如下：</p>
            <!-- 验证码 -->
            <table width="100%%" cellpadding="0" cellspacing="0">
              <tr>
                <td align="center" style="padding: 20px 0;">
                  <div style="background-color: #f8f9fa; border: 2px dashed #667eea; border-radius: 8px; padding: 20px 40px; display: inline-block;">
                    <span style="font-size: 36px; font-weight: 700; color: #667eea; letter-spacing: 8px; font-family: 'Courier New', monospace;">%s</span>
                  </div>
                </td>
              </tr>
            </table>
            <!-- 提示 -->
            <p style="color: #999; font-size: 13px; margin: 20px 0 0; text-align: center;">
              验证码 %d 分钟内有效，请勿泄露给他人。
            </p>
            <p style="color: #999; font-size: 13px; margin: 10px 0 0; text-align: center;">
              如非本人操作，请忽略此邮件。
            </p>
          </td>
        </tr>
        <!-- 底部 -->
        <tr>
          <td style="background-color: #f8f9fa; padding: 20px 30px; text-align: center; border-top: 1px solid #eee;">
            <p style="color: #aaa; font-size: 12px; margin: 0;">此邮件由系统自动发送，请勿回复</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`, siteName, purposeText, code, expireMinutes)
}
