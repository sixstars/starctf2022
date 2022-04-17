use std::io::stdin;
use std::io::stdout;
use std::io::Read;
use std::io::Write;
use std::result::Result;
use rand::Rng;
use std::fs;
use std::path::Path;
use std::os::unix::ffi::OsStrExt;

fn parse_tlv(dat: &[u8]) -> Result<(u8, usize, &[u8]), &str> {
    let mut p = 0;
    for i in 0..255 {
        //println!("{} {}",i, dat[i]);
        if dat[i] == i as u8 {
            if dat[i+1] != 0x23 {
                return Ok((dat[i+1], i, &dat[..i]));
            }
            p = i+1;
        }
    }
    if p != 0 {
        for i in p..p+0xffff {
            if (dat[i] as usize)*0x100+(dat[i+1] as usize) == i-p {
                return Ok((dat[i+1], i-p, &dat[p..i]));
            }
        }
    }
    return Err("?")
    
}

fn do_enc(inp: &mut [u8], key: &mut [u8]) {
    if key.len() < 0x20 {
        return;
    }
    let mut hdat = [0u8;0x40];
    for i in 0..0x20 {
        hdat[i] = key[i]
    }
    let mut idx = 0x10;
    for i in 0..inp.len() {
        if idx == 0x10 {
            for _ in 0..0x10 {
                do_hash(&mut hdat);
            }
            idx = 0
        }
        inp[i] = inp[i]^hdat[idx];
        hdat[idx] = 0;
        idx = idx+1;
    }
    do_hash(&mut hdat);
    for i in 0..0x20 {
        key[i] = hdat[i]
    }
}

fn do_hash(hdat: &mut [u8]) {
    if hdat.len() < 0x40 {
        return;
    }
    let mut a = u128::from_le_bytes(hdat[0..0x10].try_into().unwrap());
    let mut b = u128::from_le_bytes(hdat[0x10..0x20].try_into().unwrap());
    let mut c = u128::from_le_bytes(hdat[0x20..0x30].try_into().unwrap());
    let mut d = u128::from_le_bytes(hdat[0x30..0x40].try_into().unwrap());
    for _ in 0..64 {
        let e = (a/2+(b/3&d))^c;
        d = c;
        c = b;
        b = a;
        a = e;
    }
    let tmpa = a.to_le_bytes();
    for i in 0..0x10 {
        hdat[i] = tmpa[i];
    }
    let tmpb = b.to_le_bytes();
    for i in 0..0x10 {
        hdat[i+0x10] = tmpb[i];
    }
    let tmpc = c.to_le_bytes();
    for i in 0..0x10 {
        hdat[i+0x20] = tmpc[i];
    }
    let tmpd = d.to_le_bytes();
    for i in 0..0x10 {
        hdat[i+0x30] = tmpd[i];
    }
}

fn reset_key(k1: &[u8], k2: &[u8], kout: &mut [u8]) {
    let mk: [u8; 0x10] = [127, 112, 186, 75, 129, 242, 238, 20, 108, 196, 141, 236, 200, 6, 57, 255];
    let mut hdat = [0u8;0x40];
    for i in 0..0x20 {
        hdat[i] = k1[i];
    }
    for i in 0..0x10 {
        hdat[i+0x20] = mk[i]^0x41;
    }
    do_hash(&mut hdat);
    for i in 0..0x10 {
        hdat[i+0x30] = mk[i]^0x5e;
    }
    do_hash(&mut hdat);
    for i in 0..0x20 {
        kout[i] = (kout[i]&k2[i])^hdat[i];
    }
    
}

fn parse_inp(dat: &[u8], key: &mut [u8]) {
    if let Ok((t,l,v)) = parse_tlv(dat) {
        //println!("recv {} {} {:?}", t,l,v);
        if t==0x31 {
            if l < 0x40 {
                return;
            }
            reset_key(&v[..0x20], &v[0x20..0x40], key);
            //println!("key: {:?}", key);
        } else if t==0x41 {
            let mut ct = vec![];
            for file in fs::read_dir(".").unwrap() {
                ct.extend(file.unwrap().path().into_os_string().as_bytes());
                ct.extend([0x81,0x82,0x83,0x84]);
            }
            do_enc(&mut ct, key);
            stdout().write(&ct);
            stdout().flush();
        } else if t==0x42 {
            //let mut fname = v.clone();
            let mut fname = [0u8;0x101];
            for i in 0..v.len() {
                fname[i] = v[i];
            }
            do_enc(&mut fname, key);
            //println!("{:?}", fname);
            //println!("{:?}", v.len());
            let s = String::from_utf8(fname.to_vec()[..v.len()].to_vec()).unwrap();
            let p = Path::new(&s);
            let mut ct = fs::read(p).unwrap();
            do_enc(&mut ct, key);
            stdout().write(&ct);
            stdout().flush();
        } 
    }
}

fn main() {
    let mut v = vec![0u8; 0x20];
    let mut rng = rand::thread_rng();
    for x in v.iter_mut() {
        *x = rng.gen();
    }

    let mut inp = [0u8; 0x101];
    loop {
        stdin().read_exact(&mut inp).expect("?");
        parse_inp(&inp, &mut v);
    }
}
