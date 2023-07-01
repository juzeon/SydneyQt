import time
# selenium 3
from selenium import webdriver
from webdriver_manager.microsoft import EdgeChromiumDriverManager
import jsons
from pathlib import Path

def hack_cookies(jfile):
    
    driver_path = EdgeChromiumDriverManager().install()
    driver = webdriver.Edge()
    driver.minimize_window()
    # Download the Edge driver
    # Create a new Edge browser instance in hidden mode
    # Navigate to the Bing chatbot
    driver.get("https://www.bing.com/chat/")
    # Export the browser cookies
    cookies = jsons.dumps(driver.get_cookies())
    print(cookies)
    with open(jfile, "w", encoding="utf-8", errors="ignore") as f:
        f.write(cookies)
    # Close the browser
    driver.close()

def get_cookies_loop(jfile):
    x = 0
    cookies = False
    while x < 5 and not cookies:
        x+=1
        try:
            hack_cookies(jfile)
        except Exception as e:
            print(str(e))
        if cookies:
            x = 6 
            break



if __name__ == "__main__":
    get_cookies_loop(Path("cookies.json"))
