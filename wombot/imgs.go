package main

var Images = map[string][]string{
	"attacks": []string{
		"AgACAgIAAxkBAAIJM2GhMlHe13mG_r7cNM5CQoicbV7eAAJMtzEbWW4ISYiUsqofisjSAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ4mGiKVoSrwoFN9yVzk1csZTqVhBAAAJLuDEbl2QQSUKGDcY-VQTsAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ5GGiKcO_FNNGNab15IGglajqJDgYAAJMuDEbl2QQSYK_PxjNHzrCAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ5WGiKd6rijLJCdSnhc8NPCE1U0VnAAJNuDEbl2QQSYWZcReTIHHEAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ5mGiKh7s9P2gavf1G5A7wKz722ZaAAJPuDEbl2QQSRvA0EUvcJ0JAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ52GiKkz_GzMpBID_hmZ9MOzcd2OTAAJQuDEbl2QQSXMOwEi31HTOAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ6GGiKoyKnfvZ1Dq-AXFzyRVaV6OeAAJRuDEbl2QQScjIRtLdbxv7AQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ6WGiKrg7I6vNkVURd5NS8vMAARkGpgACUrgxG5dkEElE69tttYag7AEAAwIAA3MAAyIE",
	},
	"memes": []string{},
	"about": []string{
		"AgACAgIAAxkBAAIJ-WGj0cb-aCM4qikXXkCvFnZU43QpAAK9tzEbl2QYSTrEDio_1C8EAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ-mGj0kJ0GXpJiALs0PkXqRhyZnMyAAK-tzEbl2QYSe6CSHPOHVjdAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ-2Gj0y2-aAoBdZkNhhktH50ndkqCAALDtzEbl2QYSe_8dax8WkNhAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ_GGj06KmwIkNVPn-OFbndTkKryvhAALFtzEbl2QYSU12LgijvySoAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ_mGj09f6adIKlda7p_ZWJ7v3Xz58AALMtzEbl2QYSaA3E2QtVQNGAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ_2Gj1AjarZWzyfP9nWACyma05R_ZAALNtzEbl2QYScef2bBPWNnVAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKAAFho9QkV9wj_LdHrSL_5fiRt9-RBwACzrcxG5dkGEkgsB9-ez0mkwEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIKAWGj1ExKV0puYYPKJfreTQEBiNYsAALPtzEbl2QYSVfD0gkMleoJAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKAmGj1HMbdY9JMP_uT8Syg1gdrIRUAALQtzEbl2QYSXutjOeU8oQ1AQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKA2Gj1LT3WWT5AUh_kKM6GWrYCS3vAALRtzEbl2QYSSJzOE8ineoCAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKBGGj1R3mM63_FAtjsWApc6U0tWvUAALStzEbl2QYSfSar3k1iqOlAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKBWGj1VICB1BF8xxWeS1G5GdZKsW4AALTtzEbl2QYSWiMbeI_fncBAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKBmGj1X5HI4arhGxJnwhGAqcs-_VKAALUtzEbl2QYSSt_Im1bC9WCAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKB2Gj1ZZn-btWnvTgVBPdemlmTXdvAALVtzEbl2QYSRezxtoGCAJ-AQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKCGGj1dR4QQYpmipst6HidtR2sQ4hAALWtzEbl2QYSbnjsqbCN0zpAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKCmGj1hRfKwwDjNlwcvp37Y4z1cXqAALXtzEbl2QYSYVIySSVw3b3AQADAgADcwADIgQ",
	},
	"leps": []string{
		"AgACAgIAAxkBAAIJOWGhUFIojs9x1Gp9sxn-hj3OV8GOAAKYtzEbWW4ISQHBxMHvh4UJAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJOmGhUH3C0BIrCgABcvr2MOymvt0wyAACmrcxG1luCElsyaHrvcjAmwEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJXGGhUcFjWx9A4tq8oqVizu0UaLipAAKctzEbWW4ISQABKylr6LOnVwEAAwIAA3MAAyIE",
	},
	"qwess": []string{
		"AgACAgIAAxkBAAIJt2GiHSmp8qdV3e0koZD-q7K8RbOqAALotzEbWW4QSU0HVWntm-P_AQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJuGGiHk20JxnY4CucsiLhtP-_rDeaAALrtzEbWW4QSbMi8_lE8vgsAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJuWGiHndnz8BsG6VrbGi55qjN033LAALstzEbWW4QSXipRPSzxGeLAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJumGiHpSV4W7WG1D1N3caykNl4oFDAALttzEbWW4QSSVZO1k8mAXyAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJu2GiHskfawwNjBw-sFh2eE_RJ024AALutzEbWW4QSSSV58M9XWuQAQADAgADс",
		"AgACAgIAAxkBAAIJvGGiHvRFAAElsORhmoRuoF5AtMEJQwAC77cxG1luEEnbB4MhfbNsnQEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJvWGiHxpTggELtbq2mIiJeFpXlMNiAALwtzEbWW4QSYE-NCkraM-TAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJvmGiHy4KajdWjst_ymy_s399gq3cAALxtzEbWW4QScb1fgABRfix8QEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJv2GiH0d__iIbny4sieWPRvV2EmwDAALytzEbWW4QSR99CNJs1U_dAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJwGGiH2RjP6P-zuLt4mqdVnpm49CYAALztzEbWW4QSb31mtmBz6HJAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJwWGiH4TgCsAfyW-WOTGwIraDgffyAAL0tzEbWW4QSZa0ldBtd6MqAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJwmGiH6SSur4AAd1nvJyPIR8gGY_j_wAC9bcxG1luEEneqbg_AAH--kIBAAMCAANzAAMiBA",
		"AgACAgIAAxkBAAIJw2GiH8OiMq7Uj_zmTaKjbr-Cv8O-AAL2tzEbWW4QST-aD6T3F3UlAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJxGGiH9k6Fni9GlIcwSJ3mz2_U9mQAAL3tzEbWW4QSfGfzKkwwiRpAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJxWGiH_aMz6Yc85vc5HatskaPbnZBAAL4tzEbWW4QSTAZ0iAq-7NuAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJxmGiIBR7X4VwGw0mYNSxsW7SLXukAAL5tzEbWW4QSfzypDgvRZkgAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJx2GiICWT94t6IUwCWzWsHx5QUgABRwAC-rcxG1luEElDJrnnzjVMuAEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJyGGiID8ieIIqr6tpFc7_L6ySeMv7AAL7tzEbWW4QSf7Kkh1LGoEhAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJyWGiIFxi1sjDtA0VZjiiQvGS2i3NAAL9tzEbWW4QSYJIkwuzhwx-AQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJymGiIHQ5l1CnKCpN8wrbXsn903T6AALttzEbWW4QSSVZO1k8mAXyAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJy2GiIIi1P54AAb7gxncbHl9wu80NGQAC_rcxG1luEEmkW9QfcUvHfwEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJzGGiI_7fPNpI2-CjxWIMZvuXxr6HAAIEuDEbl2QQSfZh21N0nmNUAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJzWGiJCISw62YafqUOdhHAUfqZ3c9AAIFuDEbl2QQSd_Kzu8Fb-jRAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJzmGiJECYHLC7UWgixStACZigQ5emAAIGuDEbl2QQSeCr9petT4wvAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJz2GiJGsWQx_XaAsFHPTeZcabQAr0AAL3tzEbWW4QSfGfzKkwwiRpAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ0GGiJJ8AAbqVZ-MDIVd_ecp-jnDVgwACCrgxG5dkEEl7QuTPjPYO0QEAAwIAA3MAAyIE",
		"AgACAgIAAxkBAAIJ0WGiJM4itKkWDOpDvI1li5LEIuWyAAILuDEbl2QQSTPY1xHlk3oRAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ0mGiJOoeYnIw4mtUd6r4C61FLfMTAAIMuDEbl2QQSarZGfaMw56MAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ02GiJQ3cv-fb2fwbEzDd7mdkYitXAAINuDEbl2QQSWE_Wg332n0UAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ1GGiJSLlKv1wFIu6mApWuyem4WroAAIOuDEbl2QQSRe2WuwRIhggAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ1WGiJTldhtru69Ilsj8yTapEar9lAAIPuDEbl2QQSZsIEITCFX1GAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ4GGiJw9rqyuOqlnwVdK4xKvW0I6xAAJDuDEbl2QQSYFKecdVdO2bAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIJ4WGiJy3e9XKxf0bmtfunvKSAQ0ixAAJGuDEbl2QQST69EDRUqNY8AQADAgADcwADIgQ",
	},
	"sleep": []string{
		"AgACAgIAAxkBAAIKLWGlKBgsb84o9t51lUqd0ZqqiIa1AAIHuDEbtxMpSUi0-iqGqZYlAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKLmGlKxCr8uuKfiXnQCPTaQJsoKeCAAIMuDEbtxMpSXlKzhqWHK6BAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKL2GlKzshDsDmKUkYWwGx_36qLFYhAAINuDEbtxMpSWoMVqGBu5XtAQADAgADcwADIgQ",
	},
	"unsleep": []string{
		"AgACAgIAAxkBAAIKImGlJjt5XESmzUkuJlC6hczJ40trAAIBuDEbtxMpSccd77B8ExTIAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKI2GlJotLWBx0lkDxxyjOP6C3yzZGAAIDuDEbtxMpSduyTo1hPzlHAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKJGGlJ0CCfbFsEB1TOiyD9FDeSnPWAAIEuDEbtxMpSYzMTLtgPAunAQADAgADcwADIgQ",
	},
	"kill": []string{
		"AgACAgIAAxkBAAIKNWGlLF1G3OsPCZAsIlyO3eg1-n1CAAIRuDEbtxMpSQgQFNBVFgZZAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKNmGlLM_fm9bv-AZYEyzhRsiy0sfRAAITuDEbtxMpSbt2wrC9eHsfAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKOGGlLcP9CX3fYBROkXKB-vsPIEeSAAIYuDEbtxMpSbZa_iISVYScAQADAgADcwADIgQ",
	},
	"schweine": []string{
		"AgACAgIAAxkBAAIKO2Gmjn8q_luYTUz8Sp4erSb9j4fBAAK5uTEbMuo5SUwnIdJf8mwgAQADAgADcwADIgQ",
	},
	"new": []string{
		"AgACAgIAAxkBAAIKPWGmkMpgjmvloT95w1nIOVAxp_l1AAK7uTEbMuo5Sei6jrqQeGNIAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKPmGmkSR6SxhffiIh1K_kkS-1h1qfAAK8uTEbMuo5SbLlCJ276iAjAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKP2GmkULEiaEYtyuTV9WVwq2MNPiyAAK9uTEbMuo5Sc9BU1wW-5ZDAQADAgADcwADIgQ",
		"AgACAgIAAxkBAAIKQGGmkXK16CAyeqdFpCPuBUCEm222AAK-uTEbMuo5SZ9AV4vFiprqAQADAgADcwADIgQ",
	},
	"cancel_0": []string{
		"AgACAgIAAxkBAAIKQWGns-vqVhHXg7y-5JgcAq11I4cYAAJrujEbMupBSVLysU_UEmCCAQADAgADcwADIgQ",
	},
	"cancel_1": []string{
		"AgACAgIAAxkBAAIKUGGntgYkXMFbeSBlIIlC6g7aI8tMAAJtujEbMupBSSwv6Ej9v6YjAQADAgADcwADIgQ",
	},
	"laughter": []string{
		"AgACAgIAAxkBAAI9rmLz7xGCEExlSac_db8qIgeN0hJDAAJzwDEb4HegS__XJKoPeZxQAQADAgADbQADKQQ",
		"AgACAgIAAxkBAAI9sWLz70uwxeED9r6aDHcdTnMlBOxeAAKbwDEb4HegSyh2GzwEIyzeAQADAgADeAADKQQ",
	},
}