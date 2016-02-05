/*$Rev$*/
DELETE FROM releaseregex where ID < 100000;
INSERT INTO `releaseregex` (`ID`, `groupname`, `regex`, `ordinal`, `status`, `description`, `categoryID`) VALUES (631, 'alt.binaries.worms', '/^\\[u4all.*?\\](?P<name>.*?(xvid|x264|bluray).*?)\\[(?P<parts>\\d{1,3}\\/\\d{1,3})\\]/', 0, 1, '', NULL);
INSERT INTO `releaseregex` (`ID`, `groupname`, `regex`, `ordinal`, `status`, `description`, `categoryID`) VALUES (627, 'alt.binaries.multimedia', '/^#A.*?: (?P<name>.*?) \\[(?P<parts>\\d{1,3}\\/\\d{1,3})/i', 1, 1, '', NULL);
INSERT INTO `releaseregex` (`ID`, `groupname`, `regex`, `ordinal`, `status`, `description`, `categoryID`) VALUES (625, 'alt.binaries.multimedia', '/^#a\\..*?\\- req.*?\\- (?P<name>.*?)( \\-|) \\[(?P<parts>\\d{1,3}\\/\\d{1,3})/i', 1, 1, '', NULL);
INSERT INTO `releaseregex` (`ID`, `groupname`, `regex`, `ordinal`, `status`, `description`, `categoryID`) VALUES (626, 'alt.binaries.hdtv.x264', '/^\\[(?P<name>\\w.*?)\\]\\-\\[ich.*?\\((?P<parts>\\d{1,3}\\/\\d{1,3})/i', 1, 1, '', NULL);
INSERT INTO `releaseregex` (`ID`, `groupname`, `regex`, `ordinal`, `status`, `description`, `categoryID`) VALUES (623, 'alt.binaries.mp3', '/^(?P<name>\\w.*?\\-\\w.*?)\\[(?P<parts>\\d{1,3}\\/\\d{1,3})/i', 1, 1, '', NULL);
