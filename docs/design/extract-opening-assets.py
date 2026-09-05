from PIL import Image,ImageDraw,ImageChops
from pathlib import Path
im=Image.open('web/public/iceberg-station-mark-v2.png').convert('RGBA')
out=Path('web/public/iceberg/intro')
out.mkdir(parents=True, exist_ok=True)
# Original pixels only; polygon masks separate the existing artwork.
parts={
'boat':[(123,187),(234,180),(253,184),(268,198),(270,216),(261,244),(247,261),(236,269),(245,262),(267,260),(283,266),(293,285),(319,280),(331,285),(352,298),(411,292),(443,292),(456,298),(457,307),(447,326),(429,349),(408,371),(376,393),(345,405),(313,412),(273,415),(245,408),(218,398),(190,396),(165,386),(151,390),(136,367),(123,350),(137,345),(124,328),(115,309),(110,284),(114,258),(121,242),(132,229),(144,222),(164,212),(185,207),(205,207),(245,205),(239,201),(128,200),(126,198),(244,194),(126,191)],
'ice-left':[(42,180),(48,159),(85,133),(104,97),(130,90),(158,58),(208,30),(222,56),(227,71),(247,85),(243,117),(239,132),(255,160),(253,178),(211,180),(120,180),(64,213)],
'ice-right':[(293,215),(319,189),(319,155),(330,137),(334,111),(356,98),(374,72),(414,103),(424,132),(443,138),(459,184),(476,199),(478,233),(487,248),(490,289),(455,289),(426,289),(352,295),(329,279),(292,282)],
'splash':[(52,390),(87,385),(71,367),(61,361),(71,359),(97,369),(96,346),(106,351),(126,368),(151,384),(151,376),(171,387),(193,395),(196,387),(220,401),(237,399),(263,415),(208,415),(155,406),(105,403)],
'chip':[(258,108),(292,131),(289,154),(274,150),(257,130)]}
for name,points in parts.items():
 mask=Image.new('L',im.size);ImageDraw.Draw(mask).polygon(points,fill=255)
 layer=im.copy();layer.putalpha(ImageChops.multiply(im.getchannel('A'),mask));box=layer.getbbox();layer=layer.crop(box);layer.save(out/(name+'.png'));print(name,box,layer.size)
# Trim occluded bases so no shrimp or boat pixels remain in the ice layers.
for name, height in [('ice-left', 145), ('ice-right', 204)]:
    path = out / (name + '.png')
    image = Image.open(path)
    image.crop((0, 0, image.width, height)).save(path)
