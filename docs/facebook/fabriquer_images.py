# Fabrique les images Facebook (docs/facebook) à partir de captures du jeu (mode CAPSAVE, voir CLAUDE.md).
# Les chemins des captures sont ceux du dossier de travail de la session : à adapter.
from PIL import Image, ImageDraw, ImageFont, ImageFilter
import sys
F='/root/go/pkg/mod/github.com/golang/image@v0.20.0/font/gofont/ttfs/'
def font(n,s): return ImageFont.truetype(F+n,s)
OUT='/home/user/Toute-une-ville/docs/facebook/'
# crops (sans les boutons du jeu) : (fichier, boîte)
shots={
 'ferme':('cap9/ferme_02a.png',(560,100,1600,800)),
 'super':('cap5/super_04a.png',(0,200,1600,800)),
 'domaine':('cap5/domaine_03a.png',(150,120,1500,820)),
 'patisserie':('cap15/patisserie_0.png',(430,120,1570,600)),
 'exotique':('cap8/exotique_02a.png',(250,300,1600,800)),
 'boulangerie':('cap10/boulangerie_0.png',(430,120,1570,600)),
}
def shot(k,w,h):
    f,b=shots[k]
    im=Image.open(f).convert('RGB').crop(b)
    # remplir w×h en gardant les proportions
    r=max(w/im.width,h/im.height)
    im=im.resize((int(im.width*r)+1,int(im.height*r)+1),Image.LANCZOS)
    x=(im.width-w)//2; y=(im.height-h)//2
    return im.crop((x,y,x+w,y+h))
def shadow_text(d,xy,t,f,fill,anchor='la',sh=(0,0,0,140),o=3):
    d.text((xy[0]+o,xy[1]+o),t,font=f,fill=sh,anchor=anchor)
    d.text(xy,t,font=f,fill=fill,anchor=anchor)
def house(d,cx,cy,s):
    d.ellipse((cx-s,cy-s,cx+s,cy+s),fill=(108,184,74))
    d.polygon([(cx-0.62*s,cy+0.05*s),(cx,cy-0.62*s),(cx+0.62*s,cy+0.05*s)],fill=(192,57,43))
    d.rectangle((cx-0.45*s,cy+0.02*s,cx+0.45*s,cy+0.62*s),fill=(243,234,216))
    d.rectangle((cx-0.12*s,cy+0.25*s,cx+0.12*s,cy+0.62*s),fill=(138,90,51))
    d.rectangle((cx+0.2*s,cy+0.15*s,cx+0.38*s,cy+0.33*s),fill=(127,179,213))

# 1. publication avec lien : 1200 × 630
W,H=1200,630
img=Image.new('RGB',(W,H),(59,106,42))
g=[('ferme',0,0),('super',1,0),('patisserie',0,1),('exotique',1,1)]
cw,ch=330,315
for k,cx,cy in g:
    img.paste(shot(k,cw,ch),(W-2*cw+cx*cw,cy*ch))
d=ImageDraw.Draw(img,'RGBA')
for x in range(W-2*cw-60,W-2*cw+1):
    a=int(255*(1-(x-(W-2*cw-60))/60))
d.rectangle((0,0,W-2*cw,H),fill=(47,92,34))
house(d,130,120,62)
shadow_text(d,(215,82),"Toute",font('Go-Bold.ttf',62),(255,255,255))
shadow_text(d,(60,190),"une Ville",font('Go-Bold.ttf',74),(255,230,102))
d.text((60,300),"Construis la ville de Clairbourg :",font=font('Go-Medium.ttf',26),fill=(255,255,255))
d.text((60,336),"ferme, supermarché, école, port,",font=font('Go-Regular.ttf',24),fill=(225,240,215))
d.text((60,368),"domaine, pâtisserie…",font=font('Go-Regular.ttf',24),fill=(225,240,215))
d.rounded_rectangle((60,425,480,495),radius=18,fill=(255,210,63))
d.text((270,460),"Je cherche des testeurs !",font=font('Go-Bold.ttf',28),fill=(59,42,16),anchor='mm')
d.text((60,535),"Jeu de gestion gratuit pour Windows",font=font('Go-Medium.ttf',22),fill=(255,255,255))
d.text((60,567),"par Nathalie314",font=font('Go-Italic.ttf',20),fill=(200,225,190))
img.save(OUT+'publication.jpg',quality=90)

# 2. photo de couverture : 1640 × 624 (bande de captures, titre au milieu)
W,H=1640,624
img=Image.new('RGB',(W,H),(47,92,34))
order=['exotique','ferme','super','patisserie','boulangerie']
cw=W//len(order)
for i,k in enumerate(order):
    img.paste(shot(k,cw+2,H),(i*cw,0))
d=ImageDraw.Draw(img,'RGBA')
# bandeau central lisible (les bords de la couverture sont coupés sur téléphone)
d.rounded_rectangle((W//2-430,H//2-130,W//2+430,H//2+130),radius=36,fill=(30,58,22,215))
shadow_text(d,(W//2,H//2-45),"Toute une Ville",font('Go-Bold.ttf',92),(255,230,102),anchor='mm')
d.text((W//2,H//2+50),"Jeu de gestion gratuit · Windows · par Nathalie314",font=font('Go-Medium.ttf',30),fill=(255,255,255),anchor='mm')
img.save(OUT+'couverture.jpg',quality=90)

# 3. photo de profil : 720 × 720 (Facebook la montre ronde)
S=720
img=Image.new('RGB',(S,S),(59,106,42))
d=ImageDraw.Draw(img,'RGBA')
house(d,S//2,S//2-40,250)
d.rounded_rectangle((90,560,630,660),radius=30,fill=(255,230,102))
d.text((S//2,610),"Toute une Ville",font=font('Go-Bold.ttf',58),fill=(59,42,16),anchor='mm')
img.save(OUT+'profil.png')
print('ok')
