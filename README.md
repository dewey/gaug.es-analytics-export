# Exporter for gaug.es

This is a quick and dirty script to export your analytics data from [gaug.es](https://get.gaug.es). According to the support there’s no way to export your own data. The export in the profile actually only “exports” your email address. [Very helpful](https://annoying.technology/posts/313fc17fb3e7d2e3/)!

The export format is `csv` and is structured in the following way. This should make it possible to import it into other tools if necessary. The naming is the same as gaug.es uses internally.

## Output format

```
date,views,people
2020-07-11,11,7
2020-07-12,21,15
2020-07-13,21,16
2020-07-14,31497,25870
2020-07-15,3359,2592
2020-07-16,904,692
2020-07-17,333,245
2020-07-18,197,148
```

## Usage

Edit `run_develop.sh` and fill the two needed environment variables. Then make the file executable by running `chmod +x run_develop.sh` and execute it: `./run_develop.sh`.

You can get the values for `COOKIE` and `CSRF_TOKEN` from your browser. Just log into gaug.es, open the inspector, navigate to the “Network” tab and look at the XHR requests. They will have these two headers so you can just copy the two values from there.

![extract variables from browser requests](https://user-images.githubusercontent.com/790262/92305953-25fcb600-ef8c-11ea-88ee-b8ba9b5c9be2.jpg)

Output in your terminal should look like this:

```
gauges-export|⇒ ./run_develop.sh
command-line-arguments
ts=2020-09-05T13:16:49.836782Z caller=export.go:68 msg="exporting traffic for site" site=blog.notmyhostna.me
gauges-export|⇒ 
```


