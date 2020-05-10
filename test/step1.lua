
local pkgpath = scrdir.."steps.wpk" -- make package full file name on script directory

print ""
log "starts step 1"

-- inits new package
local pkg = wpk.new()
pkg.automime = true -- put MIME type for each file if it is not given explicit
pkg.secret = "package-private-key" -- MAC private key for cryptographic hashes of any package file
pkg.crc32 = true -- generate CRC32 Castagnoli code for each file
pkg.md5 = true -- generate MD5 hash for each file

-- pack given file with common preset
local function packfile(fname, fpath, keywords)
	pkg:putfile({name=fname, keywords=keywords, [104]="schwarzlichtbezirk"}, path.join(scrdir, "media", fname))
	log(string.format("packed %d file %s, crc=%s", pkg:gettag(fname, "fid").uint32, fname, tostring(pkg:gettag(fname, "crc32"))))
end

-- open wpk-file for write
pkg:begin(pkgpath)

-- put images with keywords and author addition tags
packfile("bounty.jpg", "beach")
packfile("img1/qarataslar.jpg", "beach;rock")
packfile("img1/claustral.jpg", "beach;rock")

log(string.format("packaged %d files on sum of %d bytes", pkg.recnum, pkg:datasize()))

-- write records table, tags table and finalize wpk-file
pkg:complete()

log "done step 1."