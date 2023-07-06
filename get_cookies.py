
from selenium.webdriver.common.by import By
from selenium import webdriver
from webdriver_manager.microsoft import EdgeChromiumDriverManager
import jsons
from pathlib import Path
import time

options = webdriver.EdgeOptions()

def headless_edge(json_file, url):
    options.add_argument("--headless")  # 添加参数使浏览器无界面运行 (Make the browser run in headless mode)
    driver = EdgeChromiumDriverManager().install()  # 安装并获取Edge浏览器驱动 (Install and get the Edge browser driver)
    driver = webdriver.Edge(options=options)  # 创建Edge浏览器对象 (Create an Edge browser object)
    cookies = None  # 初始化cookies变量 (Initialize the cookies variable)
    driver.get(url)  # 打开指定的URL地址 (Open the specified URL)
    time.sleep(5)  # 程序休眠5秒钟 (Sleep for 5 seconds)
    try:
        try_click_elements(driver)  # 调用try_click_elements函数 (Call the try_click_elements function)
    except Exception as e:
        print(str(f"{e}\n\nUnable to find chat button. Attempting to export cookies.\n"))  # 打印异常信息 (Print the exception information)
    cookies = jsons.dumps(driver.get_cookies())  # 将获取到的cookies转换成JSON格式 (Convert the obtained cookies to JSON format)
    with open(json_file, "w", encoding="utf-8", errors="ignore") as f:
        f.write(cookies)  # 将cookies写入到指定的JSON文件中 (Write the cookies to the specified JSON file)
    driver.close()  # 关闭浏览器驱动 (Close the browser driver)
    return cookies  # 返回cookies (Return the cookies)

def try_click_elements(driver):
    xpath = '//button[@id="bnp_btn_accept"]'  # 按钮元素的XPath表达式 (XPath expression of the button element)
    driver.find_element(By.XPATH, xpath).click()  # 在网页上找到指定的按钮元素并进行点击操作 (Find the specified button element on the web page and perform a click operation)
    time.sleep(2)  # 程序休眠2秒钟 (Sleep for 2 seconds)
    xpath = '//a[@id="codexPrimaryButton"]'  # 链接元素的XPath表达式 (XPath expression of the link element)
    driver.find_element(By.XPATH, xpath).click()  # 在网页上找到指定的链接元素并进行点击操作 (Find the specified link element on the web page and perform a click operation)
    if path is None:  # 判断变量path是否为None (Check if the variable path is None)
        path = Path("./bing_cookies__default.json")  # 如果path为None，则将其赋值为指定路径 (If path is None, assign it a specified path)

def grab_cookies(json_file, url):
    attempts = 0
    while attempts < 2:  # 当attempts小于2时执行循环 (Execute the loop while attempts is less than 2)
        attempts += 1  # 将attempts自增1 (Increment attempts by 1)
        try:
            cookies = headless_edge(json_file, url)  # 调用headless_edge函数获取cookies (Call the headless_edge function to get cookies)
            if cookies != None:  # 判断cookies是否不为None (Check if cookies is not None)
                # 如果成功获取到cookies，则跳出循环 (Break the loop if cookies are successfully obtained)
                return cookies
        except Exception as e:
            print(str(e))  # 打印异常信息 (Print the exception information)


if __name__ == "__main__":
    # Start the loop to grab cookies and save them to "cookies.json"
    cookie_return = grab_cookies(Path("cookies.json"), 
                                    "https://www.bing.com/chat/")
    print(str(cookie_return)) # json serialized, and has been written to path
