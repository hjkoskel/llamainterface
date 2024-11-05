/*
AI translation system

From and to English
*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Translator struct {
	Url  string
	Lang LanguageName
}

type LanguageName string

const (
	LANG_ace_Arab LanguageName = "ace_Arab" //	Arab
	LANG_ace_Latn LanguageName = "ace_Latn" //	Acehnese (Latin script)
	LANG_acm_Arab LanguageName = "acm_Arab" //	Mesopotamian Arabic
	LANG_acq_Arab LanguageName = "acq_Arab" //	Ta’izzi-Adeni Arabic
	LANG_aeb_Arab LanguageName = "aeb_Arab" //	Tunisian Arabic
	LANG_afr_Latn LanguageName = "afr_Latn" //	Afrikaans
	LANG_ajp_Arab LanguageName = "ajp_Arab" //	South Levantine Arabic
	LANG_aka_Latn LanguageName = "aka_Latn" //	Akan
	LANG_amh_Ethi LanguageName = "amh_Ethi" //	Amharic
	LANG_apc_Arab LanguageName = "apc_Arab" //	North Levantine Arabic
	LANG_arb_Arab LanguageName = "arb_Arab" //	Modern Standard Arabic
	LANG_arb_Latn LanguageName = "arb_Latn" //	Modern Standard Arabic (Romanized)
	LANG_ars_Arab LanguageName = "ars_Arab" //	Najdi Arabic
	LANG_ary_Arab LanguageName = "ary_Arab" //	Moroccan Arabic
	LANG_arz_Arab LanguageName = "arz_Arab" //	Egyptian Arabic
	LANG_asm_Beng LanguageName = "asm_Beng" //	Assamese
	LANG_ast_Latn LanguageName = "ast_Latn" //	Asturian
	LANG_awa_Deva LanguageName = "awa_Deva" //	Awadhi
	LANG_ayr_Latn LanguageName = "ayr_Latn" //	Central Aymara
	LANG_azb_Arab LanguageName = "azb_Arab" //	South Azerbaijani
	LANG_azj_Latn LanguageName = "azj_Latn" //	North Azerbaijani
	LANG_bak_Cyrl LanguageName = "bak_Cyrl" //	Bashkir
	LANG_bam_Latn LanguageName = "bam_Latn" //	Bambara
	LANG_ban_Latn LanguageName = "ban_Latn" //	Balinese
	LANG_bel_Cyrl LanguageName = "bel_Cyrl" //	Belarusian
	LANG_bem_Latn LanguageName = "bem_Latn" //	Bemba
	LANG_ben_Beng LanguageName = "ben_Beng" //	Bengali
	LANG_bho_Deva LanguageName = "bho_Deva" //	Bhojpuri
	LANG_bjn_Arab LanguageName = "bjn_Arab" //	Banjar (Arabic script)
	LANG_bjn_Latn LanguageName = "bjn_Latn" //	Banjar (Latin script)
	LANG_bod_Tibt LanguageName = "bod_Tibt" //	Standard Tibetan
	LANG_bos_Latn LanguageName = "bos_Latn" //	Bosnian
	LANG_bug_Latn LanguageName = "bug_Latn" //	Buginese
	LANG_bul_Cyrl LanguageName = "bul_Cyrl" //	Bulgarian
	LANG_cat_Latn LanguageName = "cat_Latn" //	Catalan
	LANG_ceb_Latn LanguageName = "ceb_Latn" //	Cebuano
	LANG_ces_Latn LanguageName = "ces_Latn" //	Czech
	LANG_cjk_Latn LanguageName = "cjk_Latn" //	Chokwe
	LANG_ckb_Arab LanguageName = "ckb_Arab" //	Central Kurdish
	LANG_crh_Latn LanguageName = "crh_Latn" //	Crimean Tatar
	LANG_cym_Latn LanguageName = "cym_Latn" //	Welsh
	LANG_dan_Latn LanguageName = "dan_Latn" //	Danish
	LANG_deu_Latn LanguageName = "deu_Latn" //	German
	LANG_dik_Latn LanguageName = "dik_Latn" //	Southwestern Dinka
	LANG_dyu_Latn LanguageName = "dyu_Latn" //	Dyula
	LANG_dzo_Tibt LanguageName = "dzo_Tibt" //	Dzongkha
	LANG_ell_Grek LanguageName = "ell_Grek" //	Greek
	LANG_eng_Latn LanguageName = "eng_Latn" //	English
	LANG_epo_Latn LanguageName = "epo_Latn" //	Esperanto
	LANG_est_Latn LanguageName = "est_Latn" //	Estonian
	LANG_eus_Latn LanguageName = "eus_Latn" //	Basque
	LANG_ewe_Latn LanguageName = "ewe_Latn" //	Ewe
	LANG_fao_Latn LanguageName = "fao_Latn" //	Faroese
	LANG_fij_Latn LanguageName = "fij_Latn" //	Fijian
	LANG_fin_Latn LanguageName = "fin_Latn" //	Finnish
	LANG_fon_Latn LanguageName = "fon_Latn" //	Fon
	LANG_fra_Latn LanguageName = "fra_Latn" //	French
	LANG_fur_Latn LanguageName = "fur_Latn" //	Friulian
	LANG_fuv_Latn LanguageName = "fuv_Latn" //	Nigerian Fulfulde
	LANG_gla_Latn LanguageName = "gla_Latn" //	Scottish Gaelic
	LANG_gle_Latn LanguageName = "gle_Latn" //	Irish
	LANG_glg_Latn LanguageName = "glg_Latn" //	Galician
	LANG_grn_Latn LanguageName = "grn_Latn" //	Guarani
	LANG_guj_Gujr LanguageName = "guj_Gujr" //	Gujarati
	LANG_hat_Latn LanguageName = "hat_Latn" //	Haitian Creole
	LANG_hau_Latn LanguageName = "hau_Latn" //	Hausa
	LANG_heb_Hebr LanguageName = "heb_Hebr" //	Hebrew
	LANG_hin_Deva LanguageName = "hin_Deva" //	Hindi
	LANG_hne_Deva LanguageName = "hne_Deva" //	Chhattisgarhi
	LANG_hrv_Latn LanguageName = "hrv_Latn" //	Croatian
	LANG_hun_Latn LanguageName = "hun_Latn" //	Hungarian
	LANG_hye_Armn LanguageName = "hye_Armn" //	Armenian
	LANG_ibo_Latn LanguageName = "ibo_Latn" //	Igbo
	LANG_ilo_Latn LanguageName = "ilo_Latn" //	Ilocano
	LANG_ind_Latn LanguageName = "ind_Latn" //	Indonesian
	LANG_isl_Latn LanguageName = "isl_Latn" //	Icelandic
	LANG_ita_Latn LanguageName = "ita_Latn" //	Italian
	LANG_jav_Latn LanguageName = "jav_Latn" //	Javanese
	LANG_jpn_Jpan LanguageName = "jpn_Jpan" //	Japanese
	LANG_kab_Latn LanguageName = "kab_Latn" //	Kabyle
	LANG_kac_Latn LanguageName = "kac_Latn" //	Jingpho
	LANG_kam_Latn LanguageName = "kam_Latn" //	Kamba
	LANG_kan_Knda LanguageName = "kan_Knda" //	Kannada
	LANG_kas_Arab LanguageName = "kas_Arab" //	Kashmiri (Arabic script)
	LANG_kas_Deva LanguageName = "kas_Deva" //	Kashmiri (Devanagari script)
	LANG_kat_Geor LanguageName = "kat_Geor" //	Georgian
	LANG_knc_Arab LanguageName = "knc_Arab" //	Central Kanuri (Arabic script)
	LANG_knc_Latn LanguageName = "knc_Latn" //	Central Kanuri (Latin script)
	LANG_kaz_Cyrl LanguageName = "kaz_Cyrl" //	Kazakh
	LANG_kbp_Latn LanguageName = "kbp_Latn" //	Kabiyè
	LANG_kea_Latn LanguageName = "kea_Latn" //	Kabuverdianu
	LANG_khm_Khmr LanguageName = "khm_Khmr" //	Khmer
	LANG_kik_Latn LanguageName = "kik_Latn" //	Kikuyu
	LANG_kin_Latn LanguageName = "kin_Latn" //	Kinyarwanda
	LANG_kir_Cyrl LanguageName = "kir_Cyrl" //	Kyrgyz
	LANG_kmb_Latn LanguageName = "kmb_Latn" //	Kimbundu
	LANG_kmr_Latn LanguageName = "kmr_Latn" //	Northern Kurdish
	LANG_kon_Latn LanguageName = "kon_Latn" //	Kikongo
	LANG_kor_Hang LanguageName = "kor_Hang" //	Korean
	LANG_lao_Laoo LanguageName = "lao_Laoo" //	Lao
	LANG_lij_Latn LanguageName = "lij_Latn" //	Ligurian
	LANG_lim_Latn LanguageName = "lim_Latn" //	Limburgish
	LANG_lin_Latn LanguageName = "lin_Latn" //	Lingala
	LANG_lit_Latn LanguageName = "lit_Latn" //	Lithuanian
	LANG_lmo_Latn LanguageName = "lmo_Latn" //	Lombard
	LANG_ltg_Latn LanguageName = "ltg_Latn" //	Latgalian
	LANG_ltz_Latn LanguageName = "ltz_Latn" //	Luxembourgish
	LANG_lua_Latn LanguageName = "lua_Latn" //	Luba-Kasai
	LANG_lug_Latn LanguageName = "lug_Latn" //	Ganda
	LANG_luo_Latn LanguageName = "luo_Latn" //	Luo
	LANG_lus_Latn LanguageName = "lus_Latn" //	Mizo
	LANG_lvs_Latn LanguageName = "lvs_Latn" //	Standard Latvian
	LANG_mag_Deva LanguageName = "mag_Deva" //	Magahi
	LANG_mai_Deva LanguageName = "mai_Deva" //	Maithili
	LANG_mal_Mlym LanguageName = "mal_Mlym" //	Malayalam
	LANG_mar_Deva LanguageName = "mar_Deva" //	Marathi
	LANG_min_Arab LanguageName = "min_Arab" //	Minangkabau (Arabic script)
	LANG_min_Latn LanguageName = "min_Latn" //	Minangkabau (Latin script)
	LANG_mkd_Cyrl LanguageName = "mkd_Cyrl" //	Macedonian
	LANG_plt_Latn LanguageName = "plt_Latn" //	Plateau Malagasy
	LANG_mlt_Latn LanguageName = "mlt_Latn" //	Maltese
	LANG_mni_Beng LanguageName = "mni_Beng" //	Meitei (Bengali script)
	LANG_khk_Cyrl LanguageName = "khk_Cyrl" //	Halh Mongolian
	LANG_mos_Latn LanguageName = "mos_Latn" //	Mossi
	LANG_mri_Latn LanguageName = "mri_Latn" //	Maori
	LANG_mya_Mymr LanguageName = "mya_Mymr" //	Burmese
	LANG_nld_Latn LanguageName = "nld_Latn" //	Dutch
	LANG_nno_Latn LanguageName = "nno_Latn" //	Norwegian Nynorsk
	LANG_nob_Latn LanguageName = "nob_Latn" //	Norwegian Bokmål
	LANG_npi_Deva LanguageName = "npi_Deva" //	Nepali
	LANG_nso_Latn LanguageName = "nso_Latn" //	Northern Sotho
	LANG_nus_Latn LanguageName = "nus_Latn" //	Nuer
	LANG_nya_Latn LanguageName = "nya_Latn" //	Nyanja
	LANG_oci_Latn LanguageName = "oci_Latn" //	Occitan
	LANG_gaz_Latn LanguageName = "gaz_Latn" //	West Central Oromo
	LANG_ory_Orya LanguageName = "ory_Orya" //	Odia
	LANG_pag_Latn LanguageName = "pag_Latn" //	Pangasinan
	LANG_pan_Guru LanguageName = "pan_Guru" //	Eastern Panjabi
	LANG_pap_Latn LanguageName = "pap_Latn" //	Papiamento
	LANG_pes_Arab LanguageName = "pes_Arab" //	Western Persian
	LANG_pol_Latn LanguageName = "pol_Latn" //	Polish
	LANG_por_Latn LanguageName = "por_Latn" //	Portuguese
	LANG_prs_Arab LanguageName = "prs_Arab" //	Dari
	LANG_pbt_Arab LanguageName = "pbt_Arab" //	Southern Pashto
	LANG_quy_Latn LanguageName = "quy_Latn" //	Ayacucho Quechua
	LANG_ron_Latn LanguageName = "ron_Latn" //	Romanian
	LANG_run_Latn LanguageName = "run_Latn" //	Rundi
	LANG_rus_Cyrl LanguageName = "rus_Cyrl" //	Russian
	LANG_sag_Latn LanguageName = "sag_Latn" //	Sango
	LANG_san_Deva LanguageName = "san_Deva" //	Sanskrit
	LANG_sat_Olck LanguageName = "sat_Olck" //	Santali
	LANG_scn_Latn LanguageName = "scn_Latn" //	Sicilian
	LANG_shn_Mymr LanguageName = "shn_Mymr" //	Shan
	LANG_sin_Sinh LanguageName = "sin_Sinh" //	Sinhala
	LANG_slk_Latn LanguageName = "slk_Latn" //	Slovak
	LANG_slv_Latn LanguageName = "slv_Latn" //	Slovenian
	LANG_smo_Latn LanguageName = "smo_Latn" //	Samoan
	LANG_sna_Latn LanguageName = "sna_Latn" //	Shona
	LANG_snd_Arab LanguageName = "snd_Arab" //	Sindhi
	LANG_som_Latn LanguageName = "som_Latn" //	Somali
	LANG_sot_Latn LanguageName = "sot_Latn" //	Southern Sotho
	LANG_spa_Latn LanguageName = "spa_Latn" //	Spanish
	LANG_als_Latn LanguageName = "als_Latn" //	Tosk Albanian
	LANG_srd_Latn LanguageName = "srd_Latn" //	Sardinian
	LANG_srp_Cyrl LanguageName = "srp_Cyrl" //	Serbian
	LANG_ssw_Latn LanguageName = "ssw_Latn" //	Swati
	LANG_sun_Latn LanguageName = "sun_Latn" //	Sundanese
	LANG_swe_Latn LanguageName = "swe_Latn" //	Swedish
	LANG_swh_Latn LanguageName = "swh_Latn" //	Swahili
	LANG_szl_Latn LanguageName = "szl_Latn" //	Silesian
	LANG_tam_Taml LanguageName = "tam_Taml" //	Tamil
	LANG_tat_Cyrl LanguageName = "tat_Cyrl" //	Tatar
	LANG_tel_Telu LanguageName = "tel_Telu" //	Telugu
	LANG_tgk_Cyrl LanguageName = "tgk_Cyrl" //	Tajik
	LANG_tgl_Latn LanguageName = "tgl_Latn" //	Tagalog
	LANG_tha_Thai LanguageName = "tha_Thai" //	Thai
	LANG_tir_Ethi LanguageName = "tir_Ethi" //	Tigrinya
	LANG_taq_Latn LanguageName = "taq_Latn" //	Tamasheq (Latin script)
	LANG_taq_Tfng LanguageName = "taq_Tfng" //	Tamasheq (Tifinagh script)
	LANG_tpi_Latn LanguageName = "tpi_Latn" //	Tok Pisin
	LANG_tsn_Latn LanguageName = "tsn_Latn" //	Tswana
	LANG_tso_Latn LanguageName = "tso_Latn" //	Tsonga
	LANG_tuk_Latn LanguageName = "tuk_Latn" //	Turkmen
	LANG_tum_Latn LanguageName = "tum_Latn" //	Tumbuka
	LANG_tur_Latn LanguageName = "tur_Latn" //	Turkish
	LANG_twi_Latn LanguageName = "twi_Latn" //	Twi
	LANG_tzm_Tfng LanguageName = "tzm_Tfng" //	Central Atlas Tamazight
	LANG_uig_Arab LanguageName = "uig_Arab" //	Uyghur
	LANG_ukr_Cyrl LanguageName = "ukr_Cyrl" //	Ukrainian
	LANG_umb_Latn LanguageName = "umb_Latn" //	Umbundu
	LANG_urd_Arab LanguageName = "urd_Arab" //	Urdu
	LANG_uzn_Latn LanguageName = "uzn_Latn" //	Northern Uzbek
	LANG_vec_Latn LanguageName = "vec_Latn" //	Venetian
	LANG_vie_Latn LanguageName = "vie_Latn" //	Vietnamese
	LANG_war_Latn LanguageName = "war_Latn" //	Waray
	LANG_wol_Latn LanguageName = "wol_Latn" //	Wolof
	LANG_xho_Latn LanguageName = "xho_Latn" //	Xhosa
	LANG_ydd_Hebr LanguageName = "ydd_Hebr" //	Eastern Yiddish
	LANG_yor_Latn LanguageName = "yor_Latn" //	Yoruba
	LANG_yue_Hant LanguageName = "yue_Hant" //	Yue Chinese
	LANG_zho_Hans LanguageName = "zho_Hans" //	Chinese (Simplified)
	LANG_zho_Hant LanguageName = "zho_Hant" //	Chinese (Traditional)
	LANG_zsm_Latn LanguageName = "zsm_Latn" //	Standard Malay
	LANG_zul_Latn LanguageName = "zul_Latn" //	Zulu
)

func IsSupportedLanguage(lang LanguageName) bool {
	lst := ListOfAllowedLanguages()
	for _, v := range lst {
		if strings.ToUpper(string(v)) == strings.ToUpper(string(lang)) { //Non case sensitive
			return true
		}
	}
	return false
}

func ToLanguage(s string) LanguageName {
	lst := ListOfAllowedLanguages()
	for _, v := range lst {
		if strings.ToUpper(string(v)) == strings.ToUpper(s) { //Non case sensitive
			return v
		}
	}
	return ""
}

func ListOfAllowedLanguages() []LanguageName {
	return []LanguageName{
		LANG_ace_Arab, LANG_ace_Latn, LANG_acm_Arab, LANG_acq_Arab, LANG_aeb_Arab, LANG_afr_Latn, LANG_ajp_Arab, LANG_aka_Latn, LANG_amh_Ethi, LANG_apc_Arab,
		LANG_arb_Arab, LANG_arb_Latn, LANG_ars_Arab, LANG_ary_Arab, LANG_arz_Arab, LANG_asm_Beng, LANG_ast_Latn, LANG_awa_Deva, LANG_ayr_Latn, LANG_azb_Arab,
		LANG_azj_Latn, LANG_bak_Cyrl, LANG_bam_Latn, LANG_ban_Latn, LANG_bel_Cyrl, LANG_bem_Latn, LANG_ben_Beng, LANG_bho_Deva, LANG_bjn_Arab, LANG_bjn_Latn,
		LANG_bod_Tibt, LANG_bos_Latn, LANG_bug_Latn, LANG_bul_Cyrl, LANG_cat_Latn, LANG_ceb_Latn, LANG_ces_Latn, LANG_cjk_Latn, LANG_ckb_Arab, LANG_crh_Latn,
		LANG_cym_Latn, LANG_dan_Latn, LANG_deu_Latn, LANG_dik_Latn, LANG_dyu_Latn, LANG_dzo_Tibt, LANG_ell_Grek, LANG_eng_Latn, LANG_epo_Latn, LANG_est_Latn,
		LANG_eus_Latn, LANG_ewe_Latn, LANG_fao_Latn, LANG_fij_Latn, LANG_fin_Latn, LANG_fon_Latn, LANG_fra_Latn, LANG_fur_Latn, LANG_fuv_Latn, LANG_gla_Latn,
		LANG_gle_Latn, LANG_glg_Latn, LANG_grn_Latn, LANG_guj_Gujr, LANG_hat_Latn, LANG_hau_Latn, LANG_heb_Hebr, LANG_hin_Deva, LANG_hne_Deva, LANG_hrv_Latn,
		LANG_hun_Latn, LANG_hye_Armn, LANG_ibo_Latn, LANG_ilo_Latn, LANG_ind_Latn, LANG_isl_Latn, LANG_ita_Latn, LANG_jav_Latn, LANG_jpn_Jpan, LANG_kab_Latn,
		LANG_kac_Latn, LANG_kam_Latn, LANG_kan_Knda, LANG_kas_Arab, LANG_kas_Deva, LANG_kat_Geor, LANG_knc_Arab, LANG_knc_Latn, LANG_kaz_Cyrl, LANG_kbp_Latn,
		LANG_kea_Latn, LANG_khm_Khmr, LANG_kik_Latn, LANG_kin_Latn, LANG_kir_Cyrl, LANG_kmb_Latn, LANG_kmr_Latn, LANG_kon_Latn, LANG_kor_Hang, LANG_lao_Laoo,
		LANG_lij_Latn, LANG_lim_Latn, LANG_lin_Latn, LANG_lit_Latn, LANG_lmo_Latn, LANG_ltg_Latn, LANG_ltz_Latn, LANG_lua_Latn, LANG_lug_Latn, LANG_luo_Latn,
		LANG_lus_Latn, LANG_lvs_Latn, LANG_mag_Deva, LANG_mai_Deva, LANG_mal_Mlym, LANG_mar_Deva, LANG_min_Arab, LANG_min_Latn, LANG_mkd_Cyrl, LANG_plt_Latn,
		LANG_mlt_Latn, LANG_mni_Beng, LANG_khk_Cyrl, LANG_mos_Latn, LANG_mri_Latn, LANG_mya_Mymr, LANG_nld_Latn, LANG_nno_Latn, LANG_nob_Latn, LANG_npi_Deva,
		LANG_nso_Latn, LANG_nus_Latn, LANG_nya_Latn, LANG_oci_Latn, LANG_gaz_Latn, LANG_ory_Orya, LANG_pag_Latn, LANG_pan_Guru, LANG_pap_Latn, LANG_pes_Arab,
		LANG_pol_Latn, LANG_por_Latn, LANG_prs_Arab, LANG_pbt_Arab, LANG_quy_Latn, LANG_ron_Latn, LANG_run_Latn, LANG_rus_Cyrl, LANG_sag_Latn, LANG_san_Deva,
		LANG_sat_Olck, LANG_scn_Latn, LANG_shn_Mymr, LANG_sin_Sinh, LANG_slk_Latn, LANG_slv_Latn, LANG_smo_Latn, LANG_sna_Latn, LANG_snd_Arab, LANG_som_Latn,
		LANG_sot_Latn, LANG_spa_Latn, LANG_als_Latn, LANG_srd_Latn, LANG_srp_Cyrl, LANG_ssw_Latn, LANG_sun_Latn, LANG_swe_Latn, LANG_swh_Latn, LANG_szl_Latn,
		LANG_tam_Taml, LANG_tat_Cyrl, LANG_tel_Telu, LANG_tgk_Cyrl, LANG_tgl_Latn, LANG_tha_Thai, LANG_tir_Ethi, LANG_taq_Latn, LANG_taq_Tfng, LANG_tpi_Latn,
		LANG_tsn_Latn, LANG_tso_Latn, LANG_tuk_Latn, LANG_tum_Latn, LANG_tur_Latn, LANG_twi_Latn, LANG_tzm_Tfng, LANG_uig_Arab, LANG_ukr_Cyrl, LANG_umb_Latn,
		LANG_urd_Arab, LANG_uzn_Latn, LANG_vec_Latn, LANG_vie_Latn, LANG_war_Latn, LANG_wol_Latn, LANG_xho_Latn, LANG_ydd_Hebr, LANG_yor_Latn, LANG_yue_Hant,
		LANG_zho_Hans, LANG_zho_Hant, LANG_zsm_Latn, LANG_zul_Latn}
}

func (p *Translator) Translate(input string, langFrom LanguageName, langTo LanguageName) (string, error) {
	if len(langFrom) == 0 || len(langTo) == 0 || langTo == langFrom {
		return input, nil //No need to translate
	}

	request, errRequesting := http.NewRequest("POST", p.Url+fmt.Sprintf("?source=%s&target=%s", langFrom, langTo), bytes.NewBuffer([]byte(input)))
	if errRequesting != nil {
		return "", errRequesting
	}
	request.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	client := &http.Client{}
	client.Timeout = time.Minute * 10 //TODO PARAMETRIZE OR CONSTANT
	response, errDo := client.Do(request)
	if errDo != nil {
		return "", fmt.Errorf("error while turntongue request %s", errDo.Error())
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("turntongue query returned with code %v  %s ", response.StatusCode, response.Status)
	}

	body, errBody := io.ReadAll(response.Body)
	if errBody != nil {
		return "", errBody
	}
	return string(body), nil
}

func (p *Translator) ToEnglish(input string) (string, error) {
	return p.Translate(input, p.Lang, LANG_eng_Latn)
}

func (p *Translator) FromEnglish(input string) (string, error) {
	return p.Translate(input, LANG_eng_Latn, p.Lang)
}
