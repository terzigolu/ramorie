# jbraincli - AI Agent Kullanım Rehberi

Bu rehber, bir Yapay Zeka (AI) ajanının `jbraincli` komut satırı aracını otonom bir şekilde kullanarak geliştirme görevlerini yönetmesi, bilgi depolaması ve ilerlemesini kaydetmesi için tasarlanmıştır.

## Temel Konseptler

- **Project (Proje):** Tüm çalışmaların ana kapsayıcısıdır. Her görev veya bilgi bir projeye aittir. Her zaman aktif bir proje bağlamında çalışılmalıdır.
- **Task (Görev):** Temel çalışma birimidir. Bir görevin başlığı, açıklaması, durumu (`TODO`, `IN_PROGRESS`, `COMPLETED`) ve önceliği vardır.
- **Annotation (Not):** Bir göreve eklenen ek bilgidir. Detaylı düşünceler, yapılan işlemlerin kayıtları (loglar), hata mesajları veya `elaborate` gibi komutlarla AI tarafından üretilen içerikler burada saklanır.
- **Memory (Hafıza):** Proje bağlamında bir bilgi tabanıdır. Tekrar kullanılabilecek komutlar, kod parçacıkları, konfigürasyon detayları gibi bilgileri depolamak için kullanılır.

## Anahtar Kural: ID Yönetimi

Neredeyse tüm komutlar bir **ID** ile çalışır. `create` veya `list` komutlarından dönen **kısa ID'leri (örn: `a6ba6295`)** bir sonraki komutta kullanmak için mutlaka yakala ve sakla.

**Önemli:** Bazı komutlar (`show`, `elaborate`) kısa ID ile çalışırken, bazıları (`complete` gibi) tam UUID gerektirebilir. Eğer "Invalid UUID format" hatası alırsan, `jbraincli task show <kısa-id>` komutunu kullanarak tam UUID'yi al ve komutu onunla yeniden dene.

## Temel AI Agent İş Akışları

### 1. Yeni Bir Geliştirme Görevine Başlama

Kullanıcı yeni bir hedef veya görev verdiğinde, izlenecek adımlar:

1.  **Aktif Projeyi Kontrol Et:** `jbraincli project list` komutu ile aktif bir proje olup olmadığını kontrol et. Aktif proje yoksa, kullanıcıdan `jbraincli project use <id>` ile bir proje seçmesini iste.
2.  **Görevi Oluştur (Taskify):** Kullanıcının isteğini hemen bir göreve dönüştür.
    ```bash
    jbraincli task create "Kullanıcının verdiği görev başlığı" --description "Görevin detaylı açıklaması ve hedefleri"
    ```
3.  **ID'yi Yakala:** Komutun çıktısından dönen kısa ID'yi (örn: `b0c94701`) hemen sakla. Bu ID, bu görevle ilgili tüm gelecekteki operasyonlar için kullanılacaktır.

### 2. Görevi Detaylandırma ve Planlama (Elaborate)

Bir görev oluşturulduktan hemen sonra, görevin nasıl yapılacağına dair bir plan oluşturmak için `elaborate` komutu kullanılmalıdır. Bu, ajanın "düşünme" ve "planlama" adımıdır.

1.  **Komutu Çalıştır:**
    ```bash
    jbraincli task elaborate <görev-id>
    ```
2.  **Planı İncele:** Komut başarılı olduktan sonra, AI tarafından üretilen planı ve adımları görmek için `jbraincli task show <görev-id>` komutunu çalıştır. Bu, görevi tamamlamak için izlenecek yol haritanı oluşturur.

### 3. Görev Üzerinde Çalışma

Kodlama, komut çalıştırma gibi aktif geliştirme adımları sırasında:

1.  **Görevi Başlat:** Çalışmaya başlamadan hemen önce, görevin durumunu `IN_PROGRESS` olarak güncelle. Bu, görevin aktif olarak ele alındığını belirtir.
    ```bash
    jbraincli task start <görev-id>
    ```
2.  **İlerlemeyi Not Al (Annotate):** Yaptığın her anlamlı işlemi (bir komut çalıştırmak, bir dosyayı düzenlemek, bir hata almak vb.) göreve not olarak ekle. Bu, hem ilerlemenin bir kaydını tutar hem de kullanıcıya ne yapıldığını şeffaf bir şekilde gösterir.
    ```bash
    jbraincli annotate <görev-id> "Makefile düzenlendi ve 'build' komutu eklendi."
    jbraincli annotate <görev-id> "Derleme sırasında 'redeclared function' hatası alındı. Çözülüyor."
    ```

### 4. Görevi Tamamlama

Görevin tüm gereksinimleri karşılandığında ve iş bittiğinde:

1.  **Görevi Bitir:**
    ```bash
    jbraincli task complete <görev-id>
    ```

### 5. Bilgi Yönetimi (Hafıza)

Çalışma sırasında öğrenilen veya yeniden kullanılabilecek bilgileri saklamak için:

1.  **Bilgiyi Hatırla (Remember):**
    ```bash
    jbraincli memory remember "Uygulamayı kurmak için 'make dev-install' komutu kullanılır."
    ```
2.  **Bilgiyi Geri Çağır (Recall):** Benzer bir problemle karşılaştığında veya bir bilgiye ihtiyaç duyduğunda hafızayı sorgula.
    ```bash
    jbraincli memory recall "uygulama kurulumu"
    ```

Bu rehberi takip ederek, bir AI ajanı `jbraincli`'ı verimli ve insan-gözetimli bir şekilde kullanabilir. 