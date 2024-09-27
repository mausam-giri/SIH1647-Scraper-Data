import os
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.parse import urljoin, urlencode
from DCA_Data import sellers, commodities, centres, years
import aiohttp
import asyncio

base_url = "https://dca.ceda.ashoka.edu.in/index.php/home/"
endpoint = "getcsv"

# params = {
#     'c': commodities['Atta (Wheat)'] + '%22',
#     's': centres['Delhi'],
#     't': sellers['Retail'],
#     'y': years['2021']
# }

def update_params(commodity_name, centre_name, seller_type, year):
    parse =  {
        'c': f"[%22{commodities[commodity_name]}%22]",
        's': centres[centre_name],
        't': sellers[seller_type],
        'y': years[year]
    }
    return f"c={parse['c']}&s={parse['s']}&t={parse['t']}&y={parse['y']}"

def get_url(params):
    # query_string = urlencode(params, doseq=True)
    final_url = f"{urljoin(base_url, endpoint)}?{params}"
    return final_url

def downloader(filename, url):
    
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'
    }
    try:
        resp = requests.get(url, headers=headers)
        resp.raise_for_status()
        if resp.status_code == 200:

            directory = os.path.dirname(filename)
            if not os.path.exists(directory):
                os.makedirs(directory)

            with open(f'{filename}.csv', 'wb') as file:
                file.write(resp.content)
            return f"Downloaded {filename}"
    except requests.RequestException as e:
        error_message = f"{url}, An error occurred: {e}"
        print(error_message)
        with open("failed.txt", "a") as fp:
            fp.write(f"{url}\n")
        return error_message

def write_csv(filename, url):
    file_based = [x for x in filename.split("/") if x]
    endpoint = url.split("?").pop()
    file_based.append(endpoint)

    with open("DownloadLinks.csv", "a+") as fp:
        fp.write(f"{",".join(file_based)}\n")

def download_pulses_wise():
    single_seller = list(sellers.items())[0]
    limited_commodities = list(commodities.items())[:10]
    limited_centres = list(centres.items())[:10]
    
    for seller_type, seller_id in sellers.items():
        for commodity_type, commodity_id in commodities.items():
            for centre_type, centre_id in centres.items():
                for year_type, year_id in years.items():
                    if centre_type not in ["nothing"]:
                        params = update_params(commodity_type, centre_type, seller_type, year_type)
                        url = get_url(params)
                        filename = f"./{seller_type.title()}/{commodity_type}/{centre_type}/{year_type}"
                        write_csv(filename, url)
                        print(f"[{seller_type.title()}] {commodity_type} - {centre_type} {year_type}", end="/r")
                        result = downloader(filename, url)
                        print(f"[{seller_type.title()}] {commodity_type} - {centre_type} {year_type}: {result}")
                    else:
                        pass

def download_pulses_wise_v1():
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = []
        for seller_type, seller_id in sellers.items():
            for commodity_type, commodity_id in commodities.items():
                for centre_type, centre_id in centres.items():
                    for year_type, year_id in years.items():
                        params = update_params(commodity_type, centre_type, seller_type, year_type)
                        url = get_url(params)
                        filename = f"./{seller_type.title()}/{commodity_type}/{centre_type}/{year_type}"
                        print(f"[[+] {seller_type.title()}] {commodity_type} - {centre_type} {year_type}")
                        futures.append(executor.submit(downloader, filename, url))

        for future, seller_type, commodity_type, centre_type, year_type in as_completed(futures):
            result = future.result()
            print(f"[{seller_type.title()}] {commodity_type} - {centre_type} {year_type}: {result}")

if __name__ == "__main__":
    print("Starting: Scrapper")
    download_pulses_wise()
    # asyncio.run(download_pulses_wise_v4())
